package almanack

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/flowmatic"
	"github.com/carlmjohnson/slackhook"
	"github.com/earthboundkid/resperr/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/timex"
	"github.com/spotlightpa/almanack/pkg/almlog"
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

func (svc Services) RefreshPageFromContentStore(ctx context.Context, page *db.Page) (err error) {
	defer errorx.Trace(&err)

	if !page.LastPublished.Valid {
		return
	}
	content, err := svc.ContentStore.GetFile(ctx, page.FilePath)
	if err != nil {
		return err
	}
	if err = page.FromTOML(content); err != nil {
		return err
	}
	return nil
}

func (svc Services) PopScheduledPages(ctx context.Context) (err, warning error) {
	var warnings []error
	err = svc.Tx.Begin(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
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

func (svc Services) RefreshPageFromArcStory(ctx context.Context, page *db.Page, story *db.Arc, refreshMetadata bool) (warnings []string, err error) {
	defer errorx.Trace(&err)

	var feedItem arc.FeedItem
	if err = json.Unmarshal(story.RawData, &feedItem); err != nil {
		return nil, err
	}
	body, warnings, err := ArcFeedItemToBody(ctx, svc, &feedItem)
	if err != nil {
		return nil, err
	}

	page.Body = body

	if refreshMetadata {
		fm, err := ArcFeedItemToFrontmatter(ctx, svc, &feedItem)
		if err != nil {
			return nil, err
		}
		// Update existing metadata without overwriting missing keys
		if page.Frontmatter == nil {
			page.Frontmatter = make(db.Map)
		}
		if page.LastPublished.Valid {
			delete(fm, "slug")
			delete(fm, "published")
		}
		maps.Copy(page.Frontmatter, fm)
	}

	return warnings, nil
}

func (svc Services) CreatePageFromArcSource(ctx context.Context, shared *db.SharedArticle, kind string) (warnings []string, err error) {
	defer errorx.Trace(&err)

	if shared.SourceType != "arc" {
		return nil, fmt.Errorf(
			"can't create new page for %d; wrong source type %q %q",
			shared.ID, shared.SourceType, shared.SourceID)
	}

	var feedItem arc.FeedItem
	if err = json.Unmarshal(shared.RawData, &feedItem); err != nil {
		return nil, err
	}
	body, warnings, err := ArcFeedItemToBody(ctx, svc, &feedItem)
	if err != nil {
		return nil, err
	}

	fm, err := ArcFeedItemToFrontmatter(ctx, svc, &feedItem)
	if err != nil {
		return nil, err
	}

	filepath := buildFilePath(fm, kind)

	if err = svc.createPageForSharedArticle(ctx, shared, body, fm, filepath); err != nil {
		return nil, err
	}
	return warnings, nil
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
	switch {
	case dbDoc.Metadata.Eyebrow != "" && slug != "":
		slug += "-" + stringx.SlugifyURL(dbDoc.Metadata.Eyebrow)
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

	filepath := buildFilePath(fm, kind)

	return svc.createPageForSharedArticle(ctx, shared, body, fm, filepath)
}

func (svc Services) createPageForSharedArticle(ctx context.Context, shared *db.SharedArticle, body string, fm map[string]any, filepath string) error {
	ignoreErr := false
	err := svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) (txerr error) {
		defer errorx.Trace(&txerr)
		p, txerr := q.CreatePage(ctx, db.CreatePageParams{
			FilePath:   filepath,
			SourceType: shared.SourceType,
			SourceID:   shared.SourceID,
		})
		if txerr != nil {
			// If the page already exists, just keep going
			if perr, ok := txerr.(*pgconn.PgError); ok &&
				perr.Code == "23505" && perr.ConstraintName == "page_path_key" {
				ignoreErr = true
			}
			return txerr
		}
		page, txerr := q.UpdatePage(ctx, db.UpdatePageParams{
			ID:               p.ID,
			SetFrontmatter:   true,
			Frontmatter:      fm,
			SetBody:          true,
			Body:             body,
			SetScheduleFor:   false,
			ScheduleFor:      db.NullTime,
			SetLastPublished: false,
		})
		if txerr != nil {
			return txerr
		}

		newSharedArt, txerr := q.UpdateSharedArticlePage(ctx, db.UpdateSharedArticlePageParams{
			PageID:          pgtype.Int8{Int64: page.ID, Valid: true},
			SharedArticleID: shared.ID,
		})
		if txerr != nil {
			return txerr
		}

		*shared = newSharedArt
		return nil
	})
	if !ignoreErr {
		return err
	}
	return nil
}

func buildFilePath(fm map[string]any, kind string) string {
	date := "1999-01-01"
	if t, ok := timex.Unwrap(fm["published"]); ok {
		date = timex.ToEST(t).Format("2006-01-02")
	}
	slug, _ := fm["internal-id"].(string)
	slug = cmp.Or(slug, "SPLXXX")
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
