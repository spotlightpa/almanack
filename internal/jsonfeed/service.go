package jsonfeed

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

type NewsFeed struct {
	URL string
}

func AddFlags(fl *flag.FlagSet) (nf *NewsFeed) {
	nf = new(NewsFeed)
	fl.StringVar(&nf.URL, "news-feed-url", "https://www.spotlightpa.org/feeds/full.json", "`URL` for published news feed")
	return nf
}

func (nf *NewsFeed) UpdateAppleNewsArchive(ctx context.Context, cl *http.Client, q *db.Queries) (err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	// Fetch the feed
	var source Feed
	if err = requests.
		URL(nf.URL).
		Client(cl).
		ToJSON(&source).
		Fetch(ctx); err != nil {
		return err
	}
	// Update feed archives
	data, err := json.Marshal(source.Items)
	if err != nil {
		return err
	}
	updated, err := q.UpsertNewsFeedArchives(ctx, data)
	l.InfoContext(ctx, "UpsertNewsFeedArchives", "updated_rows", updated)
	return err
}
