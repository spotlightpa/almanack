package almsvc

import (
	"context"
	"maps"
	"slices"

	"github.com/earthboundkid/errorx/v2"
	"github.com/earthboundkid/resperr/v2"
	"github.com/jackc/pgx/v5"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/gdocs"
	"github.com/spotlightpa/almanack/internal/services/google"
	"github.com/spotlightpa/almanack/internal/utils/pgxutil"
)

const MapSheetSourceType = "sheet-map"

func (svc Services) GetMapSheet(ctx context.Context, sheetID string) (sheet google.MapSheet, err error) {
	defer errorx.Trace(&err)

	if err = svc.ConfigureGoogleCert(ctx); err != nil {
		return
	}
	cl, err := svc.Gsvc.SheetsClient(ctx)
	if err != nil {
		return
	}
	sheet, err = google.SheetToMap(ctx, cl, sheetID)
	if err != nil {
		return
	}

	switch m := sheet.(type) {
	case *google.MapPage:
		if img, ok := svc.mapImage(ctx, m.ShareImagePath()); ok {
			m.ShareImageCredit = img.Credit
		}
	case *google.ScrollyMapPage:
		if m.ShareImageCredit == "" {
			if img, ok := svc.mapImage(ctx, m.ShareImagePath()); ok {
				m.ShareImageCredit = img.Credit
			}
		}
		for i := range m.Stops {
			stop := &m.Stops[i]
			if stop.Image == "" || stop.ImageRatio != "" {
				continue
			}
			if img, ok := svc.mapImage(ctx, stop.Image); ok {
				stop.ImageRatio = google.SizeRatio(int(img.Width), int(img.Height))
			}
		}
	}
	return sheet, nil
}

func (svc Services) mapImage(ctx context.Context, path string) (db.Image, bool) {
	img, err := svc.Queries.GetImageByPath(ctx, path)
	return img, err == nil
}

func MapSheetFeaturedDocID(sheet google.MapSheet) (string, error) {
	link := sheet.FeaturedDoc()
	if link == "" {
		return "", nil
	}
	id, err := gdocs.NormalizeID(link)
	if err != nil {
		return "", resperr.E{E: err, M: "Map Settings has an invalid Featured Story Document Link."}
	}
	return id, nil
}

func (svc Services) mapSheetBody(ctx context.Context, sheet google.MapSheet) (body string, err error) {
	defer errorx.Trace(&err)

	docID, err := MapSheetFeaturedDocID(sheet)
	if err != nil {
		return
	}
	var featured string
	if docID != "" {
		dbDoc, err := svc.Queries.GetGDocsByExternalIDWhereProcessed(ctx, docID)
		if err != nil {
			return "", pgxutil.NoRowsAs404(err,
				"featured story doc %q has not been processed", docID)
		}
		featured = dbDoc.ArticleMarkdown
	}
	return sheet.Body(featured), nil
}

func (svc Services) CreatePageFromMapSheet(ctx context.Context, sheetID string) (page *db.Page, err error) {
	defer errorx.Trace(&err)

	sheet, err := svc.GetMapSheet(ctx, sheetID)
	if err != nil {
		return
	}
	kind := sheet.PageKind()
	if !slices.Contains([]string{"news", "statecollege", "berks", "sponsored"}, kind) {
		return nil, resperr.E{M: "Header sheet has an invalid Section."}
	}
	fm := sheet.Frontmatter()
	if fm["internal-id"] == "" {
		return nil, resperr.E{M: "Header sheet missing Internal ID value"}
	}
	body, err := svc.mapSheetBody(ctx, sheet)
	if err != nil {
		return
	}

	page = &db.Page{
		FilePath:    buildFilePath(fm, kind),
		Frontmatter: fm,
		Body:        body,
		SourceType:  MapSheetSourceType,
		SourceID:    sheetID,
	}
	if existing, err := svc.Queries.GetPageByFilePath(ctx, page.FilePath); err == nil {
		return &existing, nil
	} else if !pgxutil.IsNotFound(err) {
		return nil, err
	}
	err = svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) error {
		return page.Save(ctx, txq, false)
	})
	if err != nil {
		return nil, err
	}
	return page, nil
}

func (svc Services) RefreshPageFromMapSheet(ctx context.Context, page *db.Page, refreshMetadata bool) (err error) {
	defer errorx.Trace(&err)

	sheet, err := svc.GetMapSheet(ctx, page.SourceID)
	if err != nil {
		return
	}
	body, err := svc.mapSheetBody(ctx, sheet)
	if err != nil {
		return
	}
	if refreshMetadata {
		if page.Frontmatter == nil {
			page.Frontmatter = make(db.Map)
		}
		fm := sheet.Frontmatter()
		delete(fm, "published")
		delete(fm, "internal-id")
		maps.DeleteFunc(fm, func(key string, v any) bool {
			s, ok := v.(string)
			return ok && s == ""
		})
		maps.Copy(page.Frontmatter, fm)
	}
	page.Body = body
	return nil
}
