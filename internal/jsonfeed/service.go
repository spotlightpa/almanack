package jsonfeed

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

type NewsFeed struct {
	URL string
}

func AddFlags(fl *flag.FlagSet) (nf *NewsFeed) {
	nf = new(NewsFeed)
	fl.StringVar(&nf.URL, "news-feed-url", "https://www.spotlightpa.org/feeds/full.json", "`URL` for published news feed (legacy, use database channels instead)")
	return nf
}

// FetchAndCache fetches the feed from the given URL and caches items in the database.
// Returns the list of external IDs that were in the feed.
func FetchAndCache(ctx context.Context, cl *http.Client, q *db.Queries, feedURL string) (externalIDs []string, err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	// Fetch the feed
	var source Feed
	if err = requests.
		URL(feedURL).
		Client(cl).
		ToJSON(&source).
		Fetch(ctx); err != nil {
		return nil, err
	}

	// Extract external IDs
	externalIDs = make([]string, len(source.Items))
	for i, item := range source.Items {
		externalIDs[i] = item.ID
	}

	// Update feed archives
	data, err := json.Marshal(source.Items)
	if err != nil {
		return nil, err
	}
	updated, err := q.UpsertNewsFeedArchives(ctx, data)
	l.InfoContext(ctx, "FetchAndCache", "feed_url", feedURL, "items", len(source.Items), "updated_rows", updated)
	return externalIDs, err
}
