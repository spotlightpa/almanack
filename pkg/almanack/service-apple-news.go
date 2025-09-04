package almanack

import (
	"context"
	"encoding/json"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/feed2anf"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (svc Services) PublishAppleNewsFeed(ctx context.Context) (err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	// Fetch the feed
	var source feed2anf.Feed
	if err = requests.
		URL(svc.ANF.NewsFeedURL).
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
	updated, err := svc.Queries.UpdateANFArchives(ctx, data)
	l.Info("PublishAppleNewsFeed", "updated_rows", updated)
	if err != nil {
		return err
	}
	// Check for unuploaded items
	// Build templates
	// Convert to ANF
	// Upload to Apple
	return nil
}
