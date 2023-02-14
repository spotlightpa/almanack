package almanack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/resperr"
	"github.com/carlmjohnson/workgroup"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/timex"
	"github.com/spotlightpa/almanack/pkg/almlog"
	"golang.org/x/exp/maps"
)

func (svc Services) PublishPage(ctx context.Context, q *db.Queries, page *db.Page) (err, warning error) {
	defer errorx.Trace(&err)

	page.SetURLPath()
	data, err := page.ToTOML()
	if err != nil {
		return
	}

	err = workgroup.DoFuncs(workgroup.MaxProcs,
		func() error {
			internalID, _ := page.Frontmatter["internal-id"].(string)
			title := stringx.First(internalID, page.FilePath)
			msg := fmt.Sprintf("Content: publishing %q", title)
			return svc.ContentStore.UpdateFile(ctx, msg, page.FilePath, []byte(data))
		}, func() error {
			_, warning = svc.Indexer.SaveObject(page.ToIndex(), ctx)
			return nil
		})
	if err != nil {
		return
	}

	p2, err := q.UpdatePage(ctx, db.UpdatePageParams{
		FilePath:         page.FilePath,
		URLPath:          page.URLPath.String,
		SetLastPublished: true,
		SetFrontmatter:   false,
		SetBody:          false,
		SetScheduleFor:   false,
		ScheduleFor:      db.NullTime,
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
	err = svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) (txerr error) {
		defer errorx.Trace(&txerr)

		pages, txerr := q.PopScheduledPages(ctx)
		if txerr != nil {
			return
		}
		var errs []error
		for _, page := range pages {
			txerr, warning = svc.PublishPage(ctx, q, &page)
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

	if _, err = svc.Indexer.SaveObject(page.ToIndex(), ctx); err != nil {
		return err
	}

	page.SetURLPath()
	newURLPath := page.URLPath.String
	if contentBefore == contentAfter && oldURLPath == newURLPath {
		return nil
	}

	l := almlog.FromContext(ctx)
	l.Info("Services.RefreshPageContents: page changed",
		"filepath", page.FilePath)

	_, err = svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
		FilePath:       page.FilePath,
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

	err = svc.Tx.Begin(ctx, pgx.TxOptions{}, func(q *db.Queries) (txerr error) {
		defer errorx.Trace(&txerr)

		if txerr = q.CreatePage(ctx, db.CreatePageParams{
			FilePath:   filepath,
			SourceType: shared.SourceType,
			SourceID:   shared.SourceID,
		}); txerr != nil {
			return txerr
		}

		page, txerr := q.UpdatePage(ctx, db.UpdatePageParams{
			FilePath:         filepath,
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
	if err != nil {
		return nil, err
	}
	return warnings, nil
}

func buildFilePath(fm map[string]any, kind string) string {
	date := "1999-01-01"
	if t, ok := timex.Unwrap(fm["published"]); ok {
		date = timex.ToEST(t).Format("2006-01-02")
	}
	slug, _ := fm["internal-id"].(string)
	slug = stringx.First(slug, "SPLXXX")
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
	return svc.SlackSocial.Post(ctx, slack.Message{
		Text: text,
		Attachments: []slack.Attachment{
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

func (svc Services) RefreshPageFromMailchimp(ctx context.Context, page *db.Page) (err error) {
	defer errorx.Trace(&err)

	id := page.ID
	archiveURL, err := svc.Queries.GetArchiveURLForPageID(ctx, id)
	if err != nil {
		return err
	}
	if archiveURL == "" {
		return resperr.New(http.StatusConflict, "no archiveURL for page %d", id)
	}

	body, err := mailchimp.ImportPage(ctx, svc.Client, archiveURL)
	if err != nil {
		return err
	}
	*page, err = svc.Queries.UpdatePageRawContent(ctx, db.UpdatePageRawContentParams{
		ID:         id,
		RawContent: body,
	})
	if err != nil {
		return err
	}
	return nil
}
