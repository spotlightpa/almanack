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

func (svc Services) PublishAppleNewsFeeds(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	channels, err := svc.Queries.ListAppleNewsChannels(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for i := range channels {
		channel := &channels[i]
		nf := &jsonfeed.NewsFeed{
			URL: channel.SourceFeedUrl,
		}
		anfsvc := &anf.Service{
			ChannelID: channel.AppleChannelID,
			Key:       channel.Key,
			Secret:    channel.Secret,
			Client:    svc.Client,
		}
		newerrs := svc.PublishAppleNewsFeed(ctx, nf, anfsvc)
		errs = append(errs, newerrs...)
	}
	return errors.Join(errs...)
}

func (svc Services) PublishAppleNewsFeed(ctx context.Context, nf *jsonfeed.NewsFeed, anfsvc *anf.Service) (errs []error) {
	l := almlog.FromContext(ctx)

	if err := nf.UpdateAppleNewsArchive(ctx, svc.Client, svc.Queries); err != nil {
		return []error{err}
	}
	// Check for unuploaded items
	newItems, err := svc.Queries.ListNewsFeedUpdates(ctx, nf.URL)
	if err != nil {
		return []error{err}
	}
	l.InfoContext(ctx, "PublishAppleNewsFeed: need uploading", "n", len(newItems))

	for i := range newItems {
		if err := svc.UploadToAppleNews(ctx, anfsvc, &newItems[i]); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (svc Services) UploadToAppleNews(ctx context.Context, anfsvc *anf.Service, newItem *db.NewsFeedItem) (err error) {
	defer errorx.Trace(&err)

	l := almlog.FromContext(ctx)
	// Convert to ANF
	art, err := anf.FromDB(newItem)
	if err != nil {
		return err
	}
	// Upload to Apple
	if newItem.AppleID == "" {
		l.InfoContext(ctx, "UploadToAppleNews: Create", "url", art.Metadata.CanonicalURL)
		res, err := anfsvc.Create(ctx, art)
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
		l.InfoContext(ctx, "UploadToAppleNews: Read", "url", art.Metadata.CanonicalURL)
		// Fetch revision ID
		res, err := anfsvc.ReadArticle(ctx, newItem.AppleID)
		if err != nil {
			err = fmt.Errorf("reading %q from Apple: %w", art.Metadata.CanonicalURL, err)
			l.ErrorContext(ctx, "error", "error", err)
			return err
		}
		// Do the update
		l.InfoContext(ctx, "UploadToAppleNews: Update", "url", art.Metadata.CanonicalURL)
		_, err = anfsvc.Update(ctx, art, newItem.AppleID, res.Data.Revision)
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
	l.InfoContext(ctx, "UploadToAppleNews: ok", "url", art.Metadata.CanonicalURL)
	return nil
}
