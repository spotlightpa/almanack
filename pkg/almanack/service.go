package almanack

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/slack"
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
	Logger
	Querier db.Querier
	ContentStore
	ImageStore
	Client      *http.Client
	SlackClient slack.Client
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

	start = time.Now()
	*dart, err = svc.Querier.GetArticle(ctx, dart.ArcID)
	svc.Logger.Printf("queried GetArticle in %v", time.Since(start))
	if err != nil {
		return err
	}
	if err = article.fromDB(*dart); err != nil {
		return err
	}

	return nil
}

func (svc Service) PopScheduledArticles(ctx context.Context, callback func([]*SpotlightPAArticle) error) error {
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
	// If the status of the article changed, fire callback then update the list
	if len(overdueArts) > 0 {
		if err := callback(overdueArts); err != nil {
			// TODO rollback
			return err
		}
	}
	return nil
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

func (svc Service) ListAvailableArcStories(ctx context.Context) (stories []ArcStory, err error) {
	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Querier.ListAvailableArticles(ctx)
	svc.Printf("ListAvailableArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}

	return storiesFromDB(dbArts)
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

func (svc Service) ListAllArcStories(ctx context.Context) (stories []ArcStory, err error) {
	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Querier.ListAllArticles(ctx)
	svc.Printf("ListAllArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}

	return storiesFromDB(dbArts)
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
		return "", err
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
