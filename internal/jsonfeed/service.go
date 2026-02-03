package jsonfeed

import (
	"context"
	"encoding/json"
	"flag"
	"net/http"

	"github.com/carlmjohnson/requests"
	"github.com/earthboundkid/errorx/v2"
	"github.com/jackc/pgx/v5/pgtype"
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

// UpdateAppleNewsArchiveForChannel fetches a feed and upserts items for a specific channel
func UpdateAppleNewsArchiveForChannel(ctx context.Context, cl *http.Client, q *db.Queries, channelID int64, feedURL string) (err error) {
	defer errorx.Trace(&err)
	l := almlog.FromContext(ctx)

	// Fetch the feed
	var source Feed
	if err = requests.
		URL(feedURL).
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
	updated, err := q.UpsertNewsFeedArchivesForChannel(ctx, db.UpsertNewsFeedArchivesForChannelParams{
		ChannelID: pgtype.Int8{Int64: channelID, Valid: true},
		Data:      data,
	})
	l.InfoContext(ctx, "UpsertNewsFeedArchivesForChannel", "channel_id", channelID, "updated_rows", updated)
	return err
}
