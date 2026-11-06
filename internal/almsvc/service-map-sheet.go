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

	for _, page := range pages {
		if perr := svc.publishMapPage(ctx, page.Slug, page.FeaturedDocLink, page.FilePath(), page.ToMarkdown); perr != nil {
			err = perr
		}
	}
	return
}

func (svc Services) SyncScrollyMapSheet(ctx context.Context, sheetID string) (err error) {
	defer errorx.Trace(&err)

	if err = svc.ConfigureGoogleCert(ctx); err != nil {
		return
	}
	cl, err := svc.Gsvc.SheetsClient(ctx)
	if err != nil {
		return
	}

	pages, err := google.SheetToScrollyMapPages(ctx, cl, sheetID)
	if err != nil {
		return
	}

	for _, page := range pages {
		if perr := svc.publishMapPage(ctx, page.Slug, page.FeaturedDocLink, page.FilePath(), page.ToMarkdown); perr != nil {
			err = perr
		}
	}
	return
}

func (svc Services) publishMapPage(ctx context.Context, slug, featuredDocLink, path string, toMarkdown func(string) (string, error)) error {
	l := almlog.FromContext(ctx)

	var featured string
	if featuredDocLink != "" {
		md, err := svc.featuredStoryMarkdown(ctx, featuredDocLink)
		if err != nil {
			l.ErrorContext(ctx, "publishMapPage: featuredStoryMarkdown", "slug", slug, "err", err)
		} else {
			featured = md
		}
	}

	content, err := toMarkdown(featured)
	if err != nil {
		l.ErrorContext(ctx, "publishMapPage: ToMarkdown", "slug", slug, "err", err)
		return err
	}
	msg := fmt.Sprintf("Maps: publish %q from sheet", slug)
	if err := svc.ContentStore.UpdateFile(ctx, msg, path, []byte(content)); err != nil {
		l.ErrorContext(ctx, "publishMapPage: UpdateFile", "slug", slug, "err", err)
		return err
	}
	l.InfoContext(ctx, "publishMapPage: published", "slug", slug, "path", path)
	return nil
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
