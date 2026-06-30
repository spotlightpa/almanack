package almsvc

import (
	"context"
	"fmt"

	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/google"
)

func (svc Services) SyncMapSheet(ctx context.Context, sheetID string) (err error) {
	defer errorx.Trace(&err)

	if err = svc.ConfigureGoogleCert(ctx); err != nil {
		return
	}
	cl, err := svc.Gsvc.SheetsClient(ctx)
	if err != nil {
		return
	}

	pages, err := google.SheetToMapPages(ctx, cl, sheetID)
	if err != nil {
		return
	}

	l := almlog.FromContext(ctx)
	for _, page := range pages {
		content := page.ToMarkdown()
		path := page.FilePath()
		msg := fmt.Sprintf("Maps: publish %q from sheet", page.Slug)
		if writeErr := svc.ContentStore.UpdateFile(ctx, msg, path, []byte(content)); writeErr != nil {
			l.ErrorContext(ctx, "SyncMapSheet: UpdateFile", "slug", page.Slug, "err", writeErr)
			err = writeErr
			continue
		}
		l.InfoContext(ctx, "SyncMapSheet: published", "slug", page.Slug, "path", path)
	}
	return
}