package almsvc

import (
	"context"
	"fmt"

	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/gdocs"
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
		var featured string
		if page.FeaturedDocLink != "" {
			var ferr error
			featured, ferr = svc.featuredStoryMarkdown(ctx, page.FeaturedDocLink)
			if ferr != nil {
				l.ErrorContext(ctx, "SyncMapSheet: featuredStoryMarkdown", "slug", page.Slug, "err", ferr)
				featured = ""
			}
		}

		content := page.ToMarkdown(featured)
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

func (svc Services) featuredStoryMarkdown(ctx context.Context, docLink string) (md string, err error) {
	defer errorx.Trace(&err)

	l := almlog.FromContext(ctx)

	id, err := gdocs.NormalizeID(docLink)
	if err != nil {
		return "", fmt.Errorf("invalid featured story doc link: %w", err)
	}

	l.InfoContext(ctx, "featuredStoryMarkdown: fetching + processing", "doc_id", id)

	newDoc, err := svc.CreateGDocsDoc(ctx, id)
	if err != nil {
		return "", fmt.Errorf("fetching featured story doc: %w", err)
	}
	if err := svc.ProcessGDocsDoc(ctx, *newDoc); err != nil {
		return "", fmt.Errorf("processing featured story doc: %w", err)
	}

	processed, err := svc.Queries.GetGDocsByExternalIDWhereProcessed(ctx, id)
	if err != nil {
		return "", fmt.Errorf("fetching processed featured story doc: %w", err)
	}
	l.InfoContext(ctx, "featuredStoryMarkdown: processed", "doc_id", id)
	return processed.ArticleMarkdown, nil
}
