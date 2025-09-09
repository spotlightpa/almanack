package almanack

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/errorx"
	"github.com/spotlightpa/almanack/internal/anf"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (svc Services) PublishAppleNewsFeed(ctx context.Context) (err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	if err := svc.NewsFeed.UpdateAppleNewsArchive(ctx, svc.Client, svc.Queries); err != nil {
		return err
	}
	// Check for unuploaded items
	newItems, err := svc.Queries.ListNewsFeedUpdates(ctx)
	if err != nil {
		return err
	}
	l.InfoContext(ctx, "PublishAppleNewsFeed: need uploading", "n", len(newItems))
	for i := range newItems {
		newItem := &newItems[i]
		// Convert to ANF
		art, err := anf.FromDB(newItem)
		if err != nil {
			return err
		}
		// Upload to Apple
		var res *anf.Response
		if newItem.AppleID == "" {
			res, err = svc.ANF.Publish(ctx, art)
			if err != nil {
				err = fmt.Errorf("publishing %q to Apple: %w", art.Metadata.CanonicalURL, err)
				l.ErrorContext(ctx, "error", "error", err)
				continue
			}
		} else {
			// TODO fetch revision ID
			res, err = svc.ANF.Update(ctx, art, newItem.AppleID, newItem.AppleRevision)
			if err != nil {
				err = fmt.Errorf("updating %q to Apple: %w", art.Metadata.CanonicalURL, err)
				l.ErrorContext(ctx, "error", "error", err)
				continue
			}
		}
		// Mark as uploaded
		_, err = svc.Queries.UpdateFeedAppleID(ctx, db.UpdateFeedAppleIDParams{
			ID:            newItem.ID,
			AppleID:       res.Data.ID,
			AppleRevision: res.Data.Revision,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
