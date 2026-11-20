package almsvc

import (
	"context"
	"errors"
	"fmt"

	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/db"
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

	for i := range pages {
		if img, ok := svc.mapImage(ctx, pages[i].ShareImagePath()); ok {
			pages[i].ShareImageCredit = img.Credit
		}
	}

	for _, page := range pages {
		err = errors.Join(err, svc.publishMapPage(ctx, page.Slug, page.FeaturedDocLink, page.FilePath(), page.ToMarkdown))
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
	for i := range pages {
		if pages[i].ShareImageCredit == "" {
			if img, ok := svc.mapImage(ctx, pages[i].ShareImagePath()); ok {
				pages[i].ShareImageCredit = img.Credit
			}
		}
		for j := range pages[i].Stops {
			stop := &pages[i].Stops[j]
			if stop.Image == "" || stop.ImageRatio != "" {
				continue
			}
			if img, ok := svc.mapImage(ctx, stop.Image); ok {
				stop.ImageRatio = google.SizeRatio(int(img.Width), int(img.Height))
			}
		}
	}

	for _, page := range pages {
		err = errors.Join(err, svc.publishMapPage(ctx, page.Slug, page.FeaturedDocLink, page.FilePath(), page.ToMarkdown))
	}
	return
}

func (svc Services) mapImage(ctx context.Context, path string) (db.Image, bool) {
	img, err := svc.Queries.GetImageByPath(ctx, path)
	return img, err == nil
}

func (svc Services) publishMapPage(ctx context.Context, slug, featuredDocLink, path string, toMarkdown func(string) (string, error)) error {
	l := almlog.FromContext(ctx)

	var featured string
	if featuredDocLink != "" {
		md, err := svc.featuredStoryMarkdown(ctx, featuredDocLink)
		if err != nil {
			l.ErrorContext(ctx, "publishMapPage: featuredStoryMarkdown", "slug", slug, "err", err)
			return err
		}
		featured = md
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
