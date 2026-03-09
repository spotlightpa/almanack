package almanack

import (
	"context"

	"github.com/carlmjohnson/flowmatic"
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
	if err = svc.Queries.ResetYouTubeMaxID(ctx); err != nil {
		return err
	}
	return flowmatic.Do(
		func() error {
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
		},
		func() error {
			data, err := svc.Queries.ListYouTubeWhereRegular(ctx, db.ListYouTubeWhereRegularParams{
				Limit:  20,
				Offset: 0,
			})
			if err != nil {
				return err
			}
			return UploadJSON(
				ctx,
				svc.FileStore,
				"feeds/youtube-regular.json",
				"public, max-age=300",
				struct {
					Videos []youtube.FeedItem `json:"videos"`
				}{
					youtube.ToFeed(data),
				},
			)
		},
	)
}
