package almanack

import (
	"context"
	"errors"
	"fmt"

	"github.com/earthboundkid/errorx/v2"
	"github.com/jackc/pgx/v5/pgtype"
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

	// Update feed archive for this channel
	if err := jsonfeed.UpdateAppleNewsArchiveForChannel(ctx, svc.Client, svc.Queries, channel.ID, channel.FeedUrl); err != nil {
		return err
	}

	// Update last synced time
	if err := svc.Queries.UpdateAppleNewsChannelLastSynced(ctx, channel.ID); err != nil {
		l.ErrorContext(ctx, "failed to update last_synced_at", "error", err)
		// Don't fail the whole operation for this
	}

	// Get items needing upload for this channel
	newItems, err := svc.Queries.ListNewsFeedUpdatesForChannel(ctx, pgInt8(channel.ID))
	if err != nil {
		return err
	}
	l.InfoContext(ctx, "publishAppleNewsFeedForChannel: need uploading", "channel", channel.Name, "n", len(newItems))

	if len(newItems) == 0 {
		return nil
	}

	// Create ANF service with this channel's credentials
	anfSvc := anf.NewForChannel(channel.ChannelID, channel.Key, channel.Secret, svc.Client)

	var errs []error
	for i := range newItems {
		err := svc.uploadToAppleNewsWithService(ctx, anfSvc, &newItems[i])
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (svc Services) uploadToAppleNewsWithService(ctx context.Context, anfSvc *anf.Service, newItem *db.NewsFeedItem) (err error) {
	defer errorx.Trace(&err)

	l := almlog.FromContext(ctx)
	// Convert to ANF
	art, err := anf.FromDB(newItem)
	if err != nil {
		return err
	}
	// Upload to Apple
	if newItem.AppleID == "" {
		l.InfoContext(ctx, "uploadToAppleNewsWithService: Create", "url", art.Metadata.CanonicalURL)
		res, err := anfSvc.Create(ctx, art)
		if err != nil {
			err = fmt.Errorf("publishing %q to Apple: %w", art.Metadata.CanonicalURL, err)
			l.ErrorContext(ctx, "error", "error", err)
			return err
		}
		// Mark as uploaded
		if _, err = svc.Queries.UpdateFeedAppleID(ctx, db.UpdateFeedAppleIDParams{
			ID:            newItem.ID,
			AppleID:       res.Data.ID,
			AppleShareUrl: res.Data.ShareURL,
		}); err != nil {
			return err
		}
	} else {
		l.InfoContext(ctx, "uploadToAppleNewsWithService: Read", "url", art.Metadata.CanonicalURL)
		// Fetch revision ID
		res, err := anfSvc.ReadArticle(ctx, newItem.AppleID)
		if err != nil {
			err = fmt.Errorf("reading %q from Apple: %w", art.Metadata.CanonicalURL, err)
			l.ErrorContext(ctx, "error", "error", err)
			return err
		}
		// Do the update
		l.InfoContext(ctx, "uploadToAppleNewsWithService: Update", "url", art.Metadata.CanonicalURL)
		_, err = anfSvc.Update(ctx, art, newItem.AppleID, res.Data.Revision)
		if err != nil {
			err = fmt.Errorf("updating %q to Apple: %w", art.Metadata.CanonicalURL, err)
			l.ErrorContext(ctx, "error", "error", err)
			return err
		}
		// Mark as uploaded
		if _, err = svc.Queries.UpdateFeedUploaded(ctx, newItem.ID); err != nil {
			return err
		}
	}
	l.InfoContext(ctx, "uploadToAppleNewsWithService: ok", "url", art.Metadata.CanonicalURL)
	return nil
}

// MigrateToAppleNewsChannelsIfNeeded creates channel ID 1 from legacy flags
// and migrates existing news_feed_items if channel 1 doesn't exist.
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

	// Point existing news_feed_items to channel 1
	migrated, err := svc.Queries.MigrateNewsFeedItemsToChannel(ctx, pgInt8(1))
	if err != nil {
		return fmt.Errorf("migrating news_feed_items: %w", err)
	}
	l.InfoContext(ctx, "MigrateToAppleNewsChannelsIfNeeded: migrated items", "count", migrated)

	return nil
}

func pgInt8(n int64) pgtype.Int8 {
	return pgtype.Int8{Int64: n, Valid: true}
}
