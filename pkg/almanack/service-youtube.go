package almanack

import (
	"context"

	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/youtube"
)

func (svc Services) UpdateYouTubeFeed(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	entries, err := svc.YT.FetchFeed(ctx, svc.Client)
	if err != nil {
		return err
	}
	if err = youtube.UpdateCache(ctx, *svc.Queries, entries); err != nil {
		return err
	}
	data, err := svc.Queries.ListYouTubeWhereShort(ctx, db.ListYouTubeWhereShortParams{
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		return err
	}
	return UploadJSON(
		ctx,
		svc.FileStore,
		"feeds/youtube-shorts.json",
		"public, max-age=300",
		struct {
			Videos []youtube.FeedItem `json:"videos"`
		}{
			youtube.ToFeed(data),
		},
	)
}
