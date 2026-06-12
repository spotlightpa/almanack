package almsvc

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/carlmjohnson/flowmatic"
	"github.com/earthboundkid/errorx/v2"
	"github.com/earthboundkid/resperr/v2"
	"github.com/earthboundkid/slackhook/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/utils/stringx"
	"github.com/spotlightpa/almanack/internal/utils/timex"
)

func (svc Services) PublishPage(ctx context.Context, txq *db.Queries, page *db.Page) (err, warning error) {
	defer errorx.Trace(&err)

	page.SetURLPath()
	data, err := page.ToTOML()
	if err != nil {
		return
	}
	// Start two goroutines.
	// In one, try the update while holding a lock.
	// If the update succeeds, also do the GitHub publish.
	// If it publishes, commit the locked update. If not, rollback.
	// In the background, do the index and issue a warning if it fails.
	// If all this goes well, swap in the db.Page to the pointer
	var p2 db.Page
	err = flowmatic.Do(
		func() (txerr error) {
			defer errorx.Trace(&txerr)

			p2, txerr = txq.UpdatePage(ctx, db.UpdatePageParams{
				ID:               page.ID,
				URLPath:          page.URLPath.String,
				SetLastPublished: true,
				SetFrontmatter:   false,
				SetBody:          false,
				SetScheduleFor:   false,
				ScheduleFor:      db.NullTime,
			})
			if txerr != nil {
				return txerr
			}

			if txerr = svc.EnsureTaxonomyPages(ctx, txq, &p2); txerr != nil {
				return txerr
			}

			internalID, _ := page.Frontmatter["internal-id"].(string)
			title := cmp.Or(internalID, page.FilePath)
			msg := fmt.Sprintf("Content: publishing %q", title)
			return svc.ContentStore.UpdateFile(ctx, msg, page.FilePath, []byte(data))
		},
		func() error {
			_, warning = svc.Indexer.SaveObject(page.ToIndex(), ctx)
			return nil
		})
	if err != nil {
		return
	}
	*page = p2
	return
}

func (svc Services) PublishJSONPage(ctx context.Context, txq *db.Queries, update db.UpdatePageParams) (page *db.Page, err error) {
	defer errorx.Trace(&err)

	// This will rollback on error
	updatedPage, err := txq.UpdatePage(ctx, update)
	if err != nil {
		return
	}

	data, err := updatedPage.ToJSON()
	if err != nil {
		return
	}

	if !update.SetLastPublished {
		return &updatedPage, nil
	}

	msg := fmt.Sprintf("Content: publishing %q", updatedPage.FilePath)
	if err = svc.ContentStore.UpdateFile(ctx, msg, updatedPage.FilePath, data); err != nil {
		return
	}

	return &updatedPage, nil
}

func (svc Services) RefreshPageFromContentStore(ctx context.Context, page *db.Page) (err error) {
	defer errorx.Trace(&err)

	if !page.LastPublished.Valid {
		return
	}
	content, err := svc.ContentStore.GetFile(ctx, page.FilePath)
	if err != nil {
		return err
	}
	if err = page.FromMD(content); err != nil {
		return err
	}
	return nil
}

func (svc Services) PopScheduledPages(ctx context.Context) (err, warning error) {
	var warnings []error
	err = svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
		defer errorx.Trace(&txerr)

		pages, txerr := txq.PopScheduledPages(ctx)
		if txerr != nil {
			return
		}
		var errs []error
		for _, page := range pages {
			txerr, warning = svc.PublishPage(ctx, txq, &page)
			errs = append(errs, txerr)
			warnings = append(warnings, warning)
		}
		return errors.Join(errs...)
	})
	return err, errors.Join(warnings...)
}

func (svc Services) EnsureTaxonomyPages(ctx context.Context, txq *db.Queries, page *db.Page) (err error) {
	var errs []error
	for _, name := range page.Series() {
		path := fmt.Sprintf("content/series/%s/_index.md", name)
		if e := svc.EnsureTaxonomyPage(ctx, path, name, txq, page); e != nil {
			errs = append(errs, e)
		}
	}
	for _, name := range page.Topics() {
		path := fmt.Sprintf("content/topics/%s/_index.md", name)
		if e := svc.EnsureTaxonomyPage(ctx, path, name, txq, page); e != nil {
			errs = append(errs, e)
		}
	}
	return errors.Join(errs...)
}

func (svc Services) EnsureTaxonomyPage(ctx context.Context, path, name string, txq *db.Queries, src *db.Page) (err error) {
	defer errorx.Trace(&err)

	// Skip if a row already exists.
	_, err = txq.GetPageByFilePath(ctx, path)
	switch {
	case err == nil:
		return nil
	case !db.IsNotFound(err):
		return err
	}

	pubDate, ok := timex.Unwrap(src.Frontmatter["published"])
	if !ok {
		pubDate = time.Now()
	}
	pubDate = timex.ToEST(pubDate)

	// Build a minimal _index.md from the source page's image-ish fields.
	index := &db.Page{
		FilePath:   path,
		SourceType: "taxonomy",
		SourceID:   src.FilePath,
		Frontmatter: db.Map{
			// "aliases": []string{},
			// "author":            "",
			// "callout-title":     "",
			// "credits":           "",
			// "dek":               "",
			"description":       src.Frontmatter["description"],
			"image":             src.Frontmatter["image"],
			"image-caption":     src.Frontmatter["image-caption"],
			"image-credit":      src.Frontmatter["image-credit"],
			"image-description": src.Frontmatter["image-description"],
			"image-gravity":     src.Frontmatter["image-gravity"],
			// "image-size":        "",
			"kicker": name,
			// "layout":            "",
			// "link":              "",
			"linktitle": cmp.Or(src.Frontmatter["blurb"], src.Frontmatter["description"]),
			"published": pubDate,
			// "related-topic":     "",
			"slug": stringx.SlugifyURL(name),
			// "subhed":            "",
			"title": name,
			// "title-tag":         "",
		}}

	data, err := index.ToTOML()
	if err != nil {
		return err
	}

	if err := index.Save(ctx, txq, true); err != nil {
		return err
	}
	msg := fmt.Sprintf("Content: publishing %q", name)
	return svc.ContentStore.UpdateFile(ctx, msg, index.FilePath, []byte(data))
}

func (svc Services) RefreshPageContents(ctx context.Context, id int64) (err error) {
	defer errorx.Trace(&err)

	page, err := svc.Queries.GetPageByID(ctx, id)
	if err != nil {
		return err
	}
	defer func(filepath string) {
		if err != nil {
			err = fmt.Errorf("problem refreshing contents of %q: %w",
				filepath, err)
		}
	}(page.FilePath)

	oldURLPath := page.URLPath.String
	contentBefore, err := page.ToTOML()
	if err != nil {
		return err
	}
	err = svc.RefreshPageFromContentStore(ctx, &page)
	if err != nil {
		return err
	}
	contentAfter, err := page.ToTOML()
	if err != nil {
		return err
	}

	page.SetURLPath()
	newURLPath := page.URLPath.String
	if contentBefore == contentAfter && oldURLPath == newURLPath {
		return nil
	}

	if _, err = svc.Indexer.SaveObject(page.ToIndex(), ctx); err != nil {
		return err
	}

	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "Services.RefreshPageContents: page changed",
		"file_path", page.FilePath, "id", page.ID)

	_, err = svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
		ID:             page.ID,
		SetFrontmatter: true,
		Frontmatter:    page.Frontmatter,
		SetBody:        true,
		Body:           page.Body,
		URLPath:        page.URLPath.String,
		ScheduleFor:    db.NullTime,
	})

	return err
}

func (svc Services) CreatePageFromGDocsDoc(ctx context.Context, shared *db.SharedArticle, kind string) (err error) {
	defer errorx.Trace(&err)

	if shared.SourceType != "gdocs" {
		return fmt.Errorf(
			"can't create new page for %d; wrong source type %q %q",
			shared.ID, shared.SourceType, shared.SourceID)
	}

	var dbDocID int64
	if err = json.Unmarshal(shared.RawData, &dbDocID); err != nil {
		return err
	}

	dbDoc, err := svc.Queries.GetGDocsByID(ctx, dbDocID)
	if !dbDoc.ProcessedAt.Valid {
		// improve
		return resperr.E{M: "Document must be processed before conversion."}
	}
	body := dbDoc.ArticleMarkdown
	slug := strings.ToLower(cmp.Or(
		dbDoc.Metadata.URLSlug,
		stringx.SlugifyURL(shared.Hed),
	))
	// Add the slugified eyebrow to the URL if it's not in there somewhere else
	eyebrowSlug := stringx.SlugifyURL(dbDoc.Metadata.Eyebrow)
	switch {
	case dbDoc.Metadata.Eyebrow != "" && slug != "" && !strings.Contains(slug, eyebrowSlug):
		slug += "-" + eyebrowSlug
	case dbDoc.Metadata.Eyebrow != "" && slug == "":
		slug = stringx.SlugifyURL(dbDoc.Metadata.Eyebrow)
	}
	fm := map[string]any{
		"internal-id":       shared.InternalID,
		"published":         shared.PublicationDate.Time,
		"byline":            shared.Byline,
		"authors":           stringx.ExtractNames(shared.Byline),
		"title":             shared.Hed,
		"description":       shared.Description,
		"blurb":             shared.Blurb,
		"image":             shared.LedeImage,
		"image-credit":      shared.LedeImageCredit,
		"image-description": shared.LedeImageDescription,
		"image-caption":     shared.LedeImageCaption,
		// Fields not exposed to Shared Admin
		"kicker":        dbDoc.Metadata.Eyebrow,
		"slug":          slug,
		"linktitle":     dbDoc.Metadata.LinkTitle,
		"title-tag":     dbDoc.Metadata.SEOTitle,
		"og-title":      dbDoc.Metadata.OGTitle,
		"twitter-title": dbDoc.Metadata.TwitterTitle,
		"layout":        dbDoc.Metadata.Layout,
	}

	if stringx.SlugifyURL(dbDoc.Metadata.Eyebrow) == "espanol" {
		fm["language-code"] = "es"
	}

	filepath := buildFilePath(fm, kind)

	return svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
		page := db.Page{
			FilePath:    filepath,
			Frontmatter: fm,
			Body:        body,
			SourceType:  shared.SourceType,
			SourceID:    shared.SourceID,
		}
		if err = page.Save(ctx, txq, false); err != nil {
			// If the page already exists, just keep going
			if db.IsUniquenessViolation(err, "page_path_key") {
				return nil
			}
			return err
		}
		newSharedArt, txerr := txq.UpdateSharedArticlePage(ctx, db.UpdateSharedArticlePageParams{
			PageID:          pgtype.Int8{Int64: page.ID, Valid: true},
			SharedArticleID: shared.ID,
		})
		if txerr != nil {
			return txerr
		}

		*shared = newSharedArt
		return txerr
	})
}

func buildFilePath(fm map[string]any, kind string) string {
	date := "1999-01-01"
	if t, ok := timex.Unwrap(fm["published"]); ok {
		date = timex.ToEST(t).Format("2006-01-02")
	}
	slug, _ := fm["internal-id"].(string)
	slug = cmp.Or(slug, "SPLXXX")
	slug, _ = stringx.StripAccents(slug)
	slug = stringx.ReplaceWhitespace(slug)

	filepath := fmt.Sprintf("content/%s/%s-%s.md", kind, date, slug)
	return filepath
}

func (svc Services) Notify(ctx context.Context, page *db.Page, publishingNow bool) (err error) {
	defer errorx.Trace(&err)

	const (
		green  = "#78bc20"
		yellow = "#ffcb05"
	)
	text := "New page publishing now…"
	color := green

	if !publishingNow {
		t := timex.ToEST(page.ScheduleFor.Time)
		text = t.Format("New article scheduled for Mon, Jan 2 at 3:04pm MST…")
		color = yellow
	}

	hed, _ := page.Frontmatter["title"].(string)
	summary := page.Frontmatter["description"].(string)
	url := page.FullURL()
	l := almlog.FromContext(ctx)
	return svc.SlackSocial.Post(ctx, l.InfoContext, svc.Client, slackhook.Message{
		Text: text,
		Attachments: []slackhook.Attachment{
			{
				Color: color,
				Fallback: fmt.Sprintf("%s\n%s\n%s",
					hed, summary, url),
				Title:     hed,
				TitleLink: url,
				Text: fmt.Sprintf(
					"%s\n%s",
					summary, url),
			},
		},
	})
}

func (svc Services) PageLoadFromContentStore(ctx context.Context, path string) (page *db.Page, err error) {
	defer errorx.Trace(&err)

	content, err := svc.ContentStore.GetFile(ctx, path)
	if err != nil {
		return nil, err
	}

	page = new(db.Page)
	if err := page.FromMD(content); err != nil {
		return nil, err
	}
	page.FilePath = path
	page.SourceType = "load"
	page.SourceID = path
	page.SetURLPath()

	err = svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) error {
		return page.Save(ctx, txq, true)
	})
	if err != nil {
		return nil, err
	}
	return page, nil
}
