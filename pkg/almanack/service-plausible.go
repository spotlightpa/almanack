package almanack

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/errorx"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (svc Services) UpdateMostPopular(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	l := almlog.FromContext(ctx)
	l.InfoCtx(ctx, "Services.UpdateMostPopular")

	pages, err := svc.Plausible.MostPopularNews(ctx, svc.Client)
	if err != nil {
		return err
	}
	if len(pages) < 5 {
		return fmt.Errorf("expected more popular pages; got %q", pages)
	}
	data, err := svc.Queries.ListPagesByURLPaths(ctx, pages)
	if err != nil {
		return err
	}
	if len(data) < 5 {
		return fmt.Errorf("could not find popular pages; got %q from %q",
			data, pages)
	}
	return UploadJSON(
		ctx,
		svc.FileStore,
		"feeds/most-popular-items.json",
		"public, max-age=300",
		struct {
			Pages []db.ListPagesByURLPathsRow `json:"pages"`
		}{
			data,
		},
	)
}
