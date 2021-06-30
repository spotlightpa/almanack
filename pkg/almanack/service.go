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
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/ganalytics"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/pkg/common"
)

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{}
}

func timeNull(nt sql.NullTime) *time.Time {
	if nt.Valid {
		t := nt.Time
		return &t
	}
	return nil
}

const timeWindow = 5 * time.Minute

func diffTime(old, new sql.NullTime) bool {
	if old.Valid != new.Valid {
		return true
	}
	if !old.Valid {
		return false
	}
	diff := old.Time.Sub(new.Time)
	if diff < 0 {
		diff = -diff
	}
	return diff > timeWindow
}

type Service struct {
	common.Logger
	Client  *http.Client
	Querier db.Querier
	common.ContentStore
	ImageStore  aws.BlobStore
	FileStore   aws.BlobStore
	SlackClient slack.Client
	Indexer     index.Indexer
	common.NewletterService
	ga *ganalytics.Client
}

func (svc Service) GetSpotlightPAArticle(ctx context.Context, dbID int32) (*SpotlightPAArticle, error) {
	start := time.Now()
	dart, err := svc.Querier.GetArticleByDBID(ctx, dbID)
	svc.Logger.Printf("queried GetArticleByDBID in %v", time.Since(start))
	if err != nil {
		return nil, db.ExpectNotFound(err)
	}
	var splArt SpotlightPAArticle
	if err = splArt.fromDB(dart); err != nil {
		return nil, err
	}
	return &splArt, nil
}

func (svc Service) GetScheduledArticle(ctx context.Context, articleID string) (*SpotlightPAArticle, error) {
	start := time.Now()
	dart, err := svc.Querier.GetArticle(ctx, nullString(articleID))
	svc.Logger.Printf("queried GetArticle in %v", time.Since(start))
	if err != nil {
		return nil, db.ExpectNotFound(err)
	}
	var schArticle SpotlightPAArticle
	if err = schArticle.fromDB(dart); err != nil {
		return nil, err
	}
	if schArticle.Empty() {
		if err = schArticle.ResetArcData(ctx, svc, dart); err != nil {
			return nil, err
		}
	} else {
		schArticle.RefreshFromContentStore(ctx, svc)
	}
	return &schArticle, nil
}

func (svc Service) ResetSpotlightPAArticleArcData(ctx context.Context, article *SpotlightPAArticle) error {
	start := time.Now()
	dart, err := svc.Querier.GetArticle(ctx, nullString(article.ArcID))
	svc.Logger.Printf("queried GetArticle in %v", time.Since(start))
	if err != nil {
		return err
	}

	return article.ResetArcData(ctx, svc, dart)
}

func (svc Service) SaveScheduledArticle(ctx context.Context, article *SpotlightPAArticle) error {
	now := time.Now()
	publishNow := false
	shouldNotify := false

	// TODO: Make less racey
	if article.ScheduleFor != nil &&
		article.ScheduleFor.Before(time.Now().Add(5*time.Minute)) {
		article.ScheduleFor = nil
		publishNow = true
	}
	article.LastSaved = &now
	dart, err := article.toDB()
	if err != nil {
		return err
	}

	start := time.Now()
	var oldSchedule sql.NullTime
	oldSchedule, err = svc.Querier.UpdateSpotlightPAArticle(ctx, db.UpdateSpotlightPAArticleParams{
		ArcID:           dart.ArcID,
		SpotlightPAPath: dart.SpotlightPAPath,
		SpotlightPAData: dart.SpotlightPAData,
		ScheduleFor:     dart.ScheduleFor,
	})
	svc.Logger.Printf("queried UpdateSpotlightPAArticle in %v", time.Since(start))
	if err != nil {
		return err
	}

	// If it was scheduled for a new time, notify
	if dart.ScheduleFor.Valid && diffTime(dart.ScheduleFor, oldSchedule) {
		shouldNotify = true
	}

	// Get the article so we can get fields not in the user article JSON
	// like filepath
	start = time.Now()
	*dart, err = svc.Querier.GetArticle(ctx, dart.ArcID)
	svc.Logger.Printf("queried GetArticle in %v", time.Since(start))
	if err != nil {
		return err
	}
	if err = article.fromDB(*dart); err != nil {
		return err
	}

	if publishNow {
		if err = article.Publish(ctx, svc); err != nil {
			// TODO rollback?
			return err
		}
		var oldTime sql.NullTime
		oldTime, err = svc.Querier.UpdateSpotlightPAArticleLastPublished(ctx, article.ArcID)
		if err != nil {
			return err
		}
		if !oldTime.Valid {
			shouldNotify = true
		}
	}

	if shouldNotify {
		// TODO: Warning only?
		if err = article.Notify(ctx, svc); err != nil {
			return err
		}
	}

	return nil
}

func (svc Service) PopScheduledArticles(ctx context.Context) error {
	start := time.Now()
	poppedArts, err := svc.Querier.PopScheduled(ctx)
	svc.Logger.Printf("queried PopScheduled in %v", time.Since(start))
	if err != nil {
		return err
	}
	overdueArts := make([]*SpotlightPAArticle, len(poppedArts))
	for i := range overdueArts {
		overdueArts[i] = new(SpotlightPAArticle)
		if err = overdueArts[i].fromDB(poppedArts[i]); err != nil {
			return err
		}
	}
	// If the status of the article changed, publish it
	var errs errutil.Slice
	for _, art := range overdueArts {
		errs.Push(art.Publish(ctx, svc))
	}
	return errs.Merge()
}

func (svc Service) GetArcStory(ctx context.Context, articleID string) (story *ArcStory, err error) {
	start := time.Now()
	dart, err := svc.Querier.GetArticle(ctx, nullString(articleID))
	svc.Printf("GetArticle query time: %v", time.Since(start))
	if err != nil {
		err = db.ExpectNotFound(err)
		return
	}
	var newstory ArcStory
	if err = newstory.fromDB(&dart); err != nil {
		return
	}
	story = &newstory
	return

}

func (svc Service) ListAvailableArcStories(ctx context.Context, page int) (stories []ArcStory, nextPage int, err error) {
	const limit = 20
	offset := int32(page) * limit

	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Querier.ListAvailableArticles(ctx, db.ListAvailableArticlesParams{
		Offset: offset,
		Limit:  limit + 1,
	})
	svc.Printf("ListAvailableArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}

	if len(dbArts) > limit {
		dbArts = dbArts[:limit]
		nextPage = page + 1
	}

	stories, err = storiesFromDB(dbArts)
	return
}

func (svc Service) SaveAlmanackArticle(ctx context.Context, article *ArcStory, setArcData bool) error {
	var (
		arcData = json.RawMessage("{}")
		err     error
	)
	if setArcData {
		if arcData, err = json.Marshal(article.ArcFeedItem); err != nil {
			return err
		}
	}
	start := time.Now()
	dart, err := svc.Querier.UpdateAlmanackArticle(ctx, db.UpdateAlmanackArticleParams{
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

func (svc Service) StoreFeed(ctx context.Context, newfeed *ArcAPI) (err error) {
	arcItems, err := json.Marshal(&newfeed.Contents)
	if err != nil {
		return err
	}
	start := time.Now()
	err = svc.Querier.UpdateArcArticles(ctx, arcItems)
	svc.Printf("StoreFeed query time: %v", time.Since(start))
	return err
}

func (svc Service) ListAllArcStories(ctx context.Context, page int) (stories []ArcStory, nextPage int, err error) {
	const limit = 50
	offset := int32(page) * limit
	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Querier.ListAllArcArticles(ctx, db.ListAllArcArticlesParams{
		Limit:  limit + 1,
		Offset: offset,
	})
	svc.Printf("ListAllArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}
	if len(dbArts) > limit {
		dbArts = dbArts[:limit]
		nextPage = page + 1
	}
	stories, err = storiesFromDB(dbArts)
	return
}

func (svc Service) ReplaceImageURL(ctx context.Context, srcURL, description, credit string) (string, error) {
	if srcURL == "" {
		return "", fmt.Errorf("no image provided")
	}
	image, err := svc.Querier.GetImageBySourceURL(ctx, srcURL)
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
	_, err = svc.Querier.CreateImage(ctx, db.CreateImageParams{
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
	data, err := json.Marshal(newItems)
	if err != nil {
		return err
	}
	if n, err := svc.Querier.UpdateNewsletterArchives(ctx, db.UpdateNewsletterArchivesParams{
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
	items, err := svc.Querier.ListNewsletters(ctx, db.ListNewslettersParams{
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
	pages, err := svc.ga.MostPopularNews(ctx)
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

	nls, err := svc.Querier.ListUnpublishedNewsletters(ctx, db.ListUnpublishedNewslettersParams{
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
	if !nl.SpotlightPAPath.Valid {
		nl.SpotlightPAPath.Valid = true
		nl.SpotlightPAPath.String = fmt.Sprintf("content/newsletters/%s/%s.md",
			nl.Type, nl.PublishedAt.Format("2006-01-02-1504"),
		)
		needsUpdate = true
	}

	// create or update the page
	if needsUpdate {
		path := nl.SpotlightPAPath.String
		if err := svc.Querier.EnsurePage(ctx, path); err != nil {
			return err
		}
		if _, err := svc.Querier.UpdatePage(ctx, db.UpdatePageParams{
			SetFrontmatter: true,
			Frontmatter: map[string]interface{}{
				"aliases":           []string{},
				"authors":           []string{},
				"blurb":             nl.Blurb,
				"byline":            "",
				"description":       nl.Description,
				"draft":             true,
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
				"slug":        "",
				"title":       nl.Subject,
				"title-tag":   "",
				"topics":      []string{},
				"url":         "",
			},
			SetBody: true,
			Body:    "",
			Path:    path,
		}); err != nil {
			return err
		}

		if nl2, err := svc.Querier.UpdateNewsletter(ctx, db.UpdateNewsletterParams{
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
	data, err := page.ToTOML()
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Content: publishing %q", page.Path)
	if err = svc.ContentStore.UpdateFile(ctx, msg, page.Path, []byte(data)); err != nil {
		return err
	}

	p2, err := svc.Querier.UpdatePage(ctx, db.UpdatePageParams{
		Path:             page.Path,
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
	content, err := svc.ContentStore.GetFile(ctx, page.Path)
	if err != nil {
		return err
	}
	if err = page.FromTOML(content); err != nil {
		return err
	}
	return nil
}
