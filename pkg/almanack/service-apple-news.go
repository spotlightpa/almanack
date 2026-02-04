package almanack

import (
	"context"
	"errors"
	"fmt"

	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/anf"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/jsonfeed"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (svc Services) PublishAppleNewsFeed(ctx context.Context) (err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	// Run one-time migration if needed (can be removed after production run)
	if err := svc.MigrateToAppleNewsChannelsIfNeeded(ctx); err != nil {
		l.ErrorContext(ctx, "MigrateToAppleNewsChannelsIfNeeded failed", "error", err)
		return err
	}

	// Get active channels from database
	channels, err := svc.Queries.ListActiveAppleNewsChannels(ctx)
	if err != nil {
		return err
	}

	if len(channels) == 0 {
		l.InfoContext(ctx, "PublishAppleNewsFeed: no active channels configured")
		return nil
	}

	l.InfoContext(ctx, "PublishAppleNewsFeed: processing channels", "n", len(channels))
	var errs []error
	for i := range channels {
		if err := svc.publishAppleNewsFeedForChannel(ctx, &channels[i]); err != nil {
			l.ErrorContext(ctx, "channel error", "channel", channels[i].Name, "error", err)
			errs = append(errs, fmt.Errorf("channel %q: %w", channels[i].Name, err))
		}
	}
	return errors.Join(errs...)
}

func (svc Services) publishAppleNewsFeedForChannel(ctx context.Context, channel *db.AppleNewsChannel) (err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	// Fetch feed and cache items (jsonfeed doesn't know about ANF)
	externalIDs, err := jsonfeed.FetchAndCache(ctx, svc.Client, svc.Queries, channel.FeedUrl)
	if err != nil {
		return err
	}

	// Update last synced time
	if err := svc.Queries.UpdateAppleNewsChannelLastSynced(ctx, channel.ID); err != nil {
		l.ErrorContext(ctx, "failed to update last_synced_at", "error", err)
		// Don't fail the whole operation for this
	}

	if len(externalIDs) == 0 {
		l.InfoContext(ctx, "publishAppleNewsFeedForChannel: no items in feed", "channel", channel.Name)
		return nil
	}

	// Find items needing upload to this channel
	newItems, err := svc.Queries.ListANFChannelItemsNeedingUpload(ctx, db.ListANFChannelItemsNeedingUploadParams{
		ExternalIds: externalIDs,
		ChannelID:   channel.ID,
	})
	if err != nil {
		return err
	}
	l.InfoContext(ctx, "publishAppleNewsFeedForChannel: need uploading", "channel", channel.Name, "n", len(newItems))

	if len(newItems) == 0 {
		return nil
	}

	// Create ANF service with this channel's credentials
	anfSvc := &anf.Service{
		ChannelID: channel.ChannelID,
		Key:       channel.Key,
		Secret:    channel.Secret,
		Client:    svc.Client,
	}

	var errs []error
	for i := range newItems {
		err := svc.uploadToAppleNewsForChannel(ctx, anfSvc, channel.ID, &newItems[i])
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (svc Services) uploadToAppleNewsForChannel(ctx context.Context, anfSvc *anf.Service, channelID int64, item *db.NewsFeedItem) (err error) {
	defer errorx.Trace(&err)

	l := almlog.FromContext(ctx)

	// Check if we already have a record for this channel+item
	existing, err := svc.Queries.GetANFChannelItem(ctx, db.GetANFChannelItemParams{
		ChannelID:      channelID,
		NewsFeedItemID: item.ID,
	})
	hasExisting := err == nil
	if err != nil && !db.IsNotFound(err) {
		return err
	}

	// Convert to ANF
	art, err := anf.FromDB(item)
	if err != nil {
		return err
	}

	var appleID, shareURL string

	if !hasExisting || existing.AppleID == "" {
		// Create new article
		l.InfoContext(ctx, "uploadToAppleNewsForChannel: Create", "url", art.Metadata.CanonicalURL)
		res, err := anfSvc.Create(ctx, art)
		if err != nil {
			err = fmt.Errorf("publishing %q to Apple: %w", art.Metadata.CanonicalURL, err)
			l.ErrorContext(ctx, "error", "error", err)
			return err
		}
		appleID = res.Data.ID
		shareURL = res.Data.ShareURL
	} else {
		// Update existing article
		l.InfoContext(ctx, "uploadToAppleNewsForChannel: Read", "url", art.Metadata.CanonicalURL)
		res, err := anfSvc.ReadArticle(ctx, existing.AppleID)
		if err != nil {
			err = fmt.Errorf("reading %q from Apple: %w", art.Metadata.CanonicalURL, err)
			l.ErrorContext(ctx, "error", "error", err)
			return err
		}

		l.InfoContext(ctx, "uploadToAppleNewsForChannel: Update", "url", art.Metadata.CanonicalURL)
		_, err = anfSvc.Update(ctx, art, existing.AppleID, res.Data.Revision)
		if err != nil {
			err = fmt.Errorf("updating %q to Apple: %w", art.Metadata.CanonicalURL, err)
			l.ErrorContext(ctx, "error", "error", err)
			return err
		}
		appleID = existing.AppleID
		shareURL = existing.AppleShareUrl
	}

	// Record the upload
	_, err = svc.Queries.UpsertANFChannelItem(ctx, db.UpsertANFChannelItemParams{
		ChannelID:      channelID,
		NewsFeedItemID: item.ID,
		AppleID:        appleID,
		AppleShareUrl:  shareURL,
	})
	if err != nil {
		return err
	}

	l.InfoContext(ctx, "uploadToAppleNewsForChannel: ok", "url", art.Metadata.CanonicalURL)
	return nil
}

// MigrateToAppleNewsChannelsIfNeeded creates channel ID 1 from legacy flags
// if channel 1 doesn't exist and legacy flags are configured.
// This function can be removed after it has run once in production.
func (svc Services) MigrateToAppleNewsChannelsIfNeeded(ctx context.Context) error {
	l := almlog.FromContext(ctx)

	// Check if channel 1 exists
	_, err := svc.Queries.GetAppleNewsChannel(ctx, 1)
	if err == nil {
		// Already migrated
		return nil
	}
	if !db.IsNotFound(err) {
		// Unexpected error
		return err
	}

	// Check if legacy flags are configured
	if svc.ANF == nil || svc.ANF.ChannelID == "" {
		l.InfoContext(ctx, "MigrateToAppleNewsChannelsIfNeeded: no legacy channel configured, skipping migration")
		return nil
	}
	if svc.NewsFeed == nil || svc.NewsFeed.URL == "" {
		l.InfoContext(ctx, "MigrateToAppleNewsChannelsIfNeeded: no legacy feed URL configured, skipping migration")
		return nil
	}

	l.InfoContext(ctx, "MigrateToAppleNewsChannelsIfNeeded: creating channel 1 from legacy flags")

	// Create channel 1 from legacy flags
	_, err = svc.Queries.CreateAppleNewsChannelWithID(ctx, db.CreateAppleNewsChannelWithIDParams{
		ID:        1,
		Name:      "Default (migrated from flags)",
		ChannelID: svc.ANF.ChannelID,
		Key:       svc.ANF.Key,
		Secret:    svc.ANF.Secret,
		FeedUrl:   svc.NewsFeed.URL,
		Active:    true,
	})
	if err != nil {
		return fmt.Errorf("creating channel 1: %w", err)
	}

	// Note: existing news_feed_item rows don't need migration since they're just a cache.
	// The anf_channel_item table starts fresh - items will be re-uploaded on next sync.
	l.InfoContext(ctx, "MigrateToAppleNewsChannelsIfNeeded: channel 1 created")

	return nil
}
