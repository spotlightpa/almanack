package almanack

import (
	"context"

	"github.com/earthboundkid/errorx/v2"
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
	return nil
}
