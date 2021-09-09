package almanack

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/resperr"
	"github.com/jackc/pgtype"
	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/internal/stringutils"
	"github.com/spotlightpa/almanack/internal/timeutil"
	"github.com/spotlightpa/almanack/pkg/common"
)

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

type Service struct {
	common.Logger
	Client  *http.Client
	Queries *db.Queries
	common.ContentStore
	ImageStore  aws.BlobStore
	FileStore   aws.BlobStore
	SlackClient slack.Client
	Indexer     index.Indexer
	common.NewletterService
	gsvc *google.Service
}

func (svc Service) GetArcStory(ctx context.Context, articleID string) (story *ArcStory, err error) {
	start := time.Now()
	dart, err := svc.Queries.GetArticleByArcID(ctx, articleID)
	svc.Printf("GetArticleByArcID query time: %v", time.Since(start))
	if err != nil {
		err = db.NoRowsAs404(err, "could not find arc-id %q", articleID)
		return
	}
	var newstory ArcStory
	if err = newstory.fromDB(&dart); err != nil {
		return
	}
	story = &newstory
	return

}

func (svc Service) ListAvailableArcStories(ctx context.Context, page int32) (stories []ArcStory, nextPage int32, err error) {
	const pageSize = 20
	offset := page * pageSize

	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Queries.ListAvailableArticles(ctx, db.ListAvailableArticlesParams{
		Offset: offset,
		Limit:  pageSize + 1,
	})
	svc.Printf("ListAvailableArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}

	if len(dbArts) > pageSize {
		dbArts = dbArts[:pageSize]
		nextPage = page + 1
	}

	stories, err = storiesFromDB(dbArts)
	return
}

func (svc Service) SaveAlmanackArticle(ctx context.Context, article *ArcStory, setArcData bool) error {
	var (
		arcData pgtype.JSONB
		err     error
	)
	if setArcData {
		if err = arcData.Set(article.FeedItem); err != nil {
			return err
		}
	}
	start := time.Now()
	dart, err := svc.Queries.UpdateAlmanackArticle(ctx, db.UpdateAlmanackArticleParams{
		ArcID:      nullString(article.ID),
		Status:     article.Status.dbstring(),
		Note:       article.Note,
		SetArcData: setArcData,
		ArcData:    arcData,
	})
	svc.Printf("UpdateAlmanackArticle query time: %v", time.Since(start))
	if err != nil {
		return err
	}
	if err = article.fromDB(&dart); err != nil {
		return err
	}

	return nil
}

func (svc Service) StoreFeed(ctx context.Context, newfeed *arc.API) (err error) {
	var arcItems pgtype.JSONB
	if err := arcItems.Set(&newfeed.Contents); err != nil {
		return err
	}
	start := time.Now()
	err = svc.Queries.UpdateArcArticles(ctx, arcItems)
	svc.Printf("StoreFeed query time: %v", time.Since(start))
	return err
}

func (svc Service) ListAllArcStories(ctx context.Context, page int32) (stories []ArcStory, nextPage int32, err error) {
	const pageSize = 50
	offset := page * pageSize
	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Queries.ListAllArcArticles(ctx, db.ListAllArcArticlesParams{
		Limit:  pageSize + 1,
		Offset: offset,
	})
	svc.Printf("ListAllArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}
	if len(dbArts) > pageSize {
		dbArts = dbArts[:pageSize]
		nextPage = page + 1
	}
	stories, err = storiesFromDB(dbArts)
	return
}

func (svc Service) ReplaceImageURL(ctx context.Context, srcURL, description, credit string) (string, error) {
	if srcURL == "" {
		return "", fmt.Errorf("no image provided")
	}
	image, err := svc.Queries.GetImageBySourceURL(ctx, srcURL)
	if err != nil && !db.IsNotFound(err) {
		return "", err
	}
	if !db.IsNotFound(err) && image.IsUploaded {
		return image.Path, nil
	}
	var path, ext string
	if path, ext, err = UploadFromURL(ctx, svc.Client, svc.ImageStore, srcURL); err != nil {
		return "", resperr.New(
			http.StatusBadGateway,
			"could not upload image %s: %w", srcURL, err,
		)
	}
	_, err = svc.Queries.CreateImage(ctx, db.CreateImageParams{
		Path:        path,
		Type:        ext,
		Description: description,
		Credit:      credit,
		SourceURL:   srcURL,
		IsUploaded:  true,
	})
	return path, err
}

func (svc Service) UpdateNewsletterArchives(ctx context.Context) error {
	return errutil.ExecParallel(
		func() error {
			return svc.UpdateNewsletterArchive(
				ctx,
				"The Investigator",
				"investigator",
				"The Investigator Newsletter Archive",
				"feeds/newsletters/investigator.json",
			)
		},
		func() error {
			return svc.UpdateNewsletterArchive(
				ctx,
				"PA Post",
				"papost",
				"PA Post Newsletter Archive",
				"feeds/newsletters/papost.json",
			)
		},
	)
}

func (svc Service) UpdateNewsletterArchive(ctx context.Context, mcType, dbType, feedtitle, path string) error {
	// get the latest from MC
	newItems, err := svc.NewletterService.ListNewletters(ctx, mcType)
	if err != nil {
		return err
	}
	// update DB
	var data pgtype.JSONB
	if err = data.Set(newItems); err != nil {
		return err
	}
	if n, err := svc.Queries.UpdateNewsletterArchives(ctx, db.UpdateNewsletterArchivesParams{
		Type: dbType,
		Data: data,
	}); err != nil {
		return err
	} else if n == 0 {
		// abort if there's nothing new to update
		svc.Logger.Printf("%q got no new items", mcType)
		return nil
	} else {
		svc.Logger.Printf("%q got %d new items", mcType, n)
	}
	// get old items list from DB
	items, err := svc.Queries.ListNewsletters(ctx, db.ListNewslettersParams{
		Type:   dbType,
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		return err
	}
	// push to S3
	feed := db.NewsletterToFeed(feedtitle, items)
	if err = UploadJSON(ctx, svc.FileStore, path, "public, max-age=300", feed); err != nil {
		return err
	}

	return nil
}

func (svc Service) UpdateMostPopular(ctx context.Context) error {
	svc.Logger.Printf("updating most popular")
	cl, err := svc.gsvc.GAClient(ctx)
	if err != nil {
		return err
	}
	pages, err := svc.gsvc.MostPopularNews(ctx, cl)
	if err != nil {
		return err
	}
	data := struct {
		Pages []string `json:"pages"`
	}{pages}
	return UploadJSON(
		ctx,
		svc.FileStore,
		"feeds/most-popular.json",
		"public, max-age=300",
		&data,
	)
}

func (svc Service) ImportNewsletterPages(ctx context.Context) (err error) {
	defer errutil.Prefix(&err, "problem importing newsletter pages")

	nls, err := svc.Queries.ListNewslettersWithoutPage(ctx, db.ListNewslettersWithoutPageParams{
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		return err
	}
	svc.Logger.Printf("importing %d newsletter pages", len(nls))
	for _, nl := range nls {
		body, err := mailchimp.ImportPage(ctx, svc.Client, nl.ArchiveURL)
		if err != nil {
			return err
		}
		if err = svc.SaveNewsletterPage(ctx, &nl, body); err != nil {
			return err
		}
	}
	return nil
}

var kickerFor = map[string]string{
	"investigator": "The Investigator",
	"papost":       "PA Post",
}

func (svc Service) SaveNewsletterPage(ctx context.Context, nl *db.Newsletter, body string) (err error) {
	defer errutil.Prefix(&err, "problem saving newsletter page")

	needsUpdate := false
	if nl.SpotlightPAPath.String == "" {
		nl.PublishedAt = timeutil.ToEST(nl.PublishedAt)
		nl.SpotlightPAPath.Valid = true
		nl.SpotlightPAPath.String = fmt.Sprintf("content/newsletters/%s/%s.md",
			nl.Type, nl.PublishedAt.Format("2006-01-02-1504"),
		)
		needsUpdate = true
	}

	// create or update the page
	if needsUpdate {
		path := nl.SpotlightPAPath.String
		if err := svc.Queries.EnsurePage(ctx, path); err != nil {
			return err
		}
		slug := stringutils.Slugify(
			timeutil.ToEST(nl.PublishedAt).Format("Jan 2 ") + nl.Subject,
		)
		if _, err := svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
			SetFrontmatter: true,
			Frontmatter: map[string]interface{}{
				"aliases":           []string{},
				"authors":           []string{},
				"blurb":             nl.Blurb,
				"byline":            "",
				"description":       nl.Description,
				"draft":             false,
				"extended-kicker":   "",
				"image":             "",
				"image-caption":     "",
				"image-credit":      "",
				"image-description": "",
				"image-size":        "",
				"internal-id": fmt.Sprintf("%s-%s",
					strings.ToUpper(nl.Type),
					nl.PublishedAt.Format("01-02-06")),
				//TODO: proper kicker lookup
				"kicker":      kickerFor[nl.Type],
				"layout":      "mailchimp-page",
				"linktitle":   "",
				"no-index":    false,
				"published":   nl.PublishedAt,
				"raw-content": body,
				"series":      []string{},
				"slug":        slug,
				"title":       nl.Subject,
				"title-tag":   "",
				"topics":      []string{},
				"url":         "",
			},
			SetBody:  true,
			Body:     "",
			FilePath: path,
		}); err != nil {
			return err
		}

		if nl2, err := svc.Queries.SetNewsletterPage(ctx, db.SetNewsletterPageParams{
			ID:              nl.ID,
			SpotlightPAPath: nl.SpotlightPAPath,
		}); err != nil {
			return err
		} else {
			*nl = nl2
		}
	}
	return nil
}

func (svc Service) PublishPage(ctx context.Context, page *db.Page) error {
	page.SetURLPath()
	data, err := page.ToTOML()
	if err != nil {
		return err
	}

	internalID, _ := page.Frontmatter["internal-id"].(string)
	title := stringutils.First(internalID, page.FilePath)
	msg := fmt.Sprintf("Content: publishing %q", title)
	if err = svc.ContentStore.UpdateFile(ctx, msg, page.FilePath, []byte(data)); err != nil {
		return err
	}
	p2, err := svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
		FilePath:         page.FilePath,
		URLPath:          page.URLPath.String,
		SetLastPublished: true,
		SetFrontmatter:   false,
		SetBody:          false,
		SetScheduleFor:   false,
	})
	if err != nil {
		return err
	}
	*page = p2
	return nil
}

func (svc Service) RefreshPageFromContentStore(ctx context.Context, page *db.Page) (err error) {
	defer errutil.Prefix(&err, "problem refreshing page content from Github")

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

func (svc Service) PopScheduledPages(ctx context.Context) error {
	pages, err := svc.Queries.PopScheduledPages(ctx)
	if err != nil {
		return err
	}

	var errs errutil.Slice
	for _, page := range pages {
		errs.Push(svc.PublishPage(ctx, &page))
	}
	return errs.Merge()
}

func (svc Service) RefreshPageContents(ctx context.Context, id int64) (err error) {
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
	})

	return err
}

func (svc Service) PageFromArcArticle(ctx context.Context, dbArt *db.Article) (page *db.Page, err error) {
	defer errutil.Prefix(&err, "problem creating page from arc")

	story, err := ArcStoryFromDB(dbArt)
	if err != nil {
		return nil, err
	}

	var splArt SpotlightPAArticle
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
	defer errutil.Prefix(&err, "problem refreshing page from arc")

	var splArt SpotlightPAArticle
	if err = story.ToArticle(ctx, svc, &splArt); err != nil {
		return err
	}

	page.Body = splArt.Body
	return nil
}

func (svc Service) Notify(ctx context.Context, page *db.Page, publishingNow bool) error {
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
	return svc.SlackClient.PostCtx(ctx, slack.Message{
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

func (svc Service) PopScheduledSiteChanges(ctx context.Context) (err error) {
	defer errutil.Trace(&err)

	configs, err := svc.Queries.PopScheduledSiteChanges(ctx, EditorsPicksLoc)
	if err != nil {
		return err
	}

	var currentConfig *db.SiteDatum
	for _, config := range configs {
		if currentConfig == nil || config.ScheduleFor.After(currentConfig.ScheduleFor) {
			currentConfig = &config
		}
	}
	if currentConfig == nil {
		svc.Printf("site data: no changes to %s", EditorsPicksLoc)
		return nil
	}
	svc.Printf("site data: updating %s", EditorsPicksLoc)

	// TODO: rollback
	if err = svc.PublishSiteConfig(ctx, currentConfig); err != nil {
		return err
	}

	return svc.Queries.CleanSiteData(ctx, EditorsPicksLoc)
}

func (svc Service) PublishSiteConfig(ctx context.Context, siteConfig *db.SiteDatum) (err error) {
	defer errutil.Trace(&err)

	data, err := json.MarshalIndent(siteConfig.Data, "", "  ")
	if err != nil {
		return err
	}
	msg := stringutils.First(MessageForLoc[siteConfig.Key], siteConfig.Key)
	if err = svc.ContentStore.UpdateFile(ctx, msg, siteConfig.Key, data); err != nil {

		return err
	}
	return nil
}
