package almanack

import (
	"context"

	"github.com/carlmjohnson/errorx"
	"github.com/spotlightpa/almanack/internal/anf"
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
		if err = svc.ANF.Publish(ctx, svc.Client, art); err != nil {
			return err
		}
		// Mark as uploaded
		_, err = svc.Queries.UpdateFeedUploaded(ctx, newItem.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
