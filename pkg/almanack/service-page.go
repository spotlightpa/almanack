package almanack

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/errutil"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/internal/stringutils"
	"github.com/spotlightpa/almanack/internal/timeutil"
)

func (svc Service) PublishPage(ctx context.Context, page *db.Page) (err, warning error) {
	defer errutil.Prefix(&err, "Service.PublishPage(%d)", page.ID)

	page.SetURLPath()
	data, err := page.ToTOML()
	if err != nil {
		return
	}

	err = errutil.ExecParallel(func() error {
		internalID, _ := page.Frontmatter["internal-id"].(string)
		title := stringutils.First(internalID, page.FilePath)
		msg := fmt.Sprintf("Content: publishing %q", title)
		return svc.ContentStore.UpdateFile(ctx, msg, page.FilePath, []byte(data))
	}, func() error {
		_, warning = svc.Indexer.SaveObject(page.ToIndex(), ctx)
		return nil
	})
	if err != nil {
		return
	}

	p2, err := svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
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

func (svc Service) RefreshPageFromContentStore(ctx context.Context, page *db.Page) (err error) {
	defer errutil.Prefix(&err, "Service.RefreshPageFromContentStore(%d)", page.ID)

	if db.IsNull(page.LastPublished) {
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

func (svc Service) PopScheduledPages(ctx context.Context) (err, warning error) {
	defer errutil.Trace(&err)

	pages, err := svc.Queries.PopScheduledPages(ctx)
	if err != nil {
		return
	}

	var errs, warnings errutil.Slice
	for _, page := range pages {
		err, warning = svc.PublishPage(ctx, &page)
		errs.Push(err)
		warnings.Push(warning)
	}
	return errs.Merge(), warnings.Merge()
}

func (svc Service) RefreshPageContents(ctx context.Context, id int64) (err error) {
	defer errutil.Trace(&err)

	page, err := svc.Queries.GetPageByID(ctx, id)
	if err != nil {
		return err
	}
	defer errutil.Prefix(&err, fmt.Sprintf("problem refreshing contents of %s", page.FilePath))

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

	svc.Printf("%s changed", page.FilePath)

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

func (svc Service) PageFromArcArticle(ctx context.Context, dbArt *db.Article, pagekind string) (page *db.Page, err error) {
	defer errutil.Trace(&err)

	story, err := ArcStoryFromDB(dbArt)
	if err != nil {
		return nil, err
	}

	var splArt SpotlightPAArticle
	splArt.PageKind = pagekind
	if err = story.ToArticle(ctx, svc, &splArt); err != nil {
		return nil, err
	}

	content, err := splArt.ToTOML()
	if err != nil {
		return nil, err
	}

	var dbPage db.Page
	if err = dbPage.FromTOML(content); err != nil {
		return nil, err
	}
	if err = svc.Queries.EnsurePage(ctx, splArt.ContentFilepath()); err != nil {
		return nil, err
	}
	if dbPage, err = svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
		FilePath:       splArt.ContentFilepath(),
		SetFrontmatter: true,
		Frontmatter:    dbPage.Frontmatter,
		SetBody:        true,
		Body:           dbPage.Body,
		URLPath:        dbPage.URLPath.String,
		ScheduleFor:    db.NullTime,
	}); err != nil {
		return nil, err
	}
	if _, err = svc.Queries.UpdateArcArticleSpotlightPAPath(ctx, db.UpdateArcArticleSpotlightPAPathParams{
		ArcID:           dbArt.ArcID.String,
		SpotlightPAPath: splArt.ContentFilepath(),
	}); err != nil {
		return nil, err
	}
	return &dbPage, nil
}

func (svc Service) RefreshPageFromArcStory(ctx context.Context, story *ArcStory, page *db.Page) (err error) {
	defer errutil.Trace(&err)

	var splArt SpotlightPAArticle
	if err = story.ToArticle(ctx, svc, &splArt); err != nil {
		return err
	}

	page.Body = splArt.Body
	return nil
}

func (svc Service) Notify(ctx context.Context, page *db.Page, publishingNow bool) (err error) {
	defer errutil.Trace(&err)

	const (
		green  = "#78bc20"
		yellow = "#ffcb05"
	)
	text := "New page publishing now…"
	color := green

	if !publishingNow {
		t := timeutil.ToEST(page.ScheduleFor.Time)
		text = t.Format("New article scheduled for Mon, Jan 2 at 3:04pm MST…")
		color = yellow
	}

	hed, _ := page.Frontmatter["title"].(string)
	summary := page.Frontmatter["description"].(string)
	url := page.FullURL()
	return svc.SlackClient.Post(ctx, slack.Message{
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
