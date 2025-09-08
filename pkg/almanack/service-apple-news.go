package almanack

import (
	"context"
	"encoding/json"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/anf"
	"github.com/spotlightpa/almanack/internal/jsonfeed"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (svc Services) PublishAppleNewsFeed(ctx context.Context) (err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	if err := svc.UpdateAppleNewsArchive(ctx); err != nil {
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
		_, err := anf.FromDB(newItem)
		if err != nil {
			return err
		}
		// TODO: Upload to Apple
		// Mark as uploaded
		_, err = svc.Queries.UpdateFeedUploaded(ctx, newItem.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (svc Services) UpdateAppleNewsArchive(ctx context.Context) (err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	// Fetch the feed
	var source jsonfeed.Feed
	if err = requests.
		URL(svc.NewsFeed.URL).
		Client(svc.Client).
		ToJSON(&source).
		Fetch(ctx); err != nil {
		return err
	}
	// Update feed archives
	data, err := json.Marshal(source.Items)
	if err != nil {
		return err
	}
	updated, err := svc.Queries.UpsertNewsFeedArchives(ctx, data)
	l.InfoContext(ctx, "UpdateAppleNewsArchive", "updated_rows", updated)
	return err
}
