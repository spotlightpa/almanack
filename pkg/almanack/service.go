package almanack

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
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

type Service struct {
	Logger
	Querier db.Querier
	ContentStore
}

func (svc Service) GetScheduledArticle(ctx context.Context, articleID string) (*ScheduledArticle, error) {
	start := time.Now()
	dart, err := svc.Querier.GetArticle(ctx, nullString(articleID))
	svc.Logger.Printf("queried GetArticle in %v", time.Since(start))
	if err != nil {
		return nil, db.ExpectNotFound(err)
	}
	var schArticle ScheduledArticle
	if err = schArticle.fromDB(dart); err != nil {
		return nil, err
	}
	return &schArticle, nil
}

func (svc Service) SaveScheduledArticle(ctx context.Context, article *ScheduledArticle) error {
	// TODO: Make less racey
	if article.ScheduleFor != nil &&
		article.ScheduleFor.Before(time.Now().Add(5*time.Minute)) {
		article.ScheduleFor = nil
		if err := article.Publish(ctx, svc.ContentStore); err != nil {
			return err
		}
	}

	// Save the article
	now := time.Now()
	article.LastSaved = &now

	dart, err := article.toDB()
	if err != nil {
		return err
	}

	start := time.Now()
	*dart, err = svc.Querier.UpdateSpotlightPAArticle(ctx, db.UpdateSpotlightPAArticleParams{
		ArcID:           dart.ArcID,
		SpotlightPAData: dart.SpotlightPAData,
		ScheduleFor:     dart.ScheduleFor,
		SpotlightPAPath: dart.SpotlightPAPath,
	})
	svc.Logger.Printf("queried UpdateSpotlightPAArticle in %v", time.Since(start))
	if err != nil {
		return err
	}

	if err = article.fromDB(*dart); err != nil {
		return err
	}
	return nil
}

func (svc Service) PopScheduledArticles(ctx context.Context, callback func([]*ScheduledArticle) error) error {
	start := time.Now()
	poppedArts, err := svc.Querier.PopScheduled(ctx)
	svc.Logger.Printf("queried PopScheduled in %v", time.Since(start))
	if err != nil {
		return err
	}
	overdueArts := make([]*ScheduledArticle, len(poppedArts))
	for i := range overdueArts {
		overdueArts[i] = new(ScheduledArticle)
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

func (svc Service) GetAvailableFeed(ctx context.Context) (stories []ArcStory, err error) {
	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Querier.ListAvailableArticles(ctx)
	svc.Printf("ListAvailableArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}
	stories = make([]ArcStory, len(dbArts))
	for i := range stories {
		if err = stories[i].fromDB(&dbArts[i]); err != nil {
			return
		}
	}
	return
}

func (svc Service) SaveAlmanackArticle(ctx context.Context, article *ArcStory) error {
	start := time.Now()
	dart, err := svc.Querier.UpdateAlmanackArticle(ctx, db.UpdateAlmanackArticleParams{
		ArcID:  nullString(article.ID),
		Status: article.Status.dbstring(),
		Note:   article.Note,
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

func (svc Service) StoreFeed(ctx context.Context, newfeed *ArcAPI, update bool) (err error) {
	arcItems, err := json.Marshal(&newfeed.Contents)
	if err != nil {
		return err
	}
	start := time.Now()
	dbarts, err := svc.Querier.UpdateArcArticles(ctx, arcItems)
	svc.Printf("StoreFeed query time: %v", time.Since(start))
	if err != nil {
		return
	}
	if update {
		newfeed.Contents = newfeed.Contents[:0]
		for i := range dbarts {
			var story ArcStory
			if err = story.fromDB(&dbarts[i]); err != nil {
				return err
			}
			newfeed.Contents = append(newfeed.Contents, story)
		}
	}
	return
}
