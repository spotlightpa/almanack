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

type Service struct {
	ContentStore
	Logger
	Querier db.Querier
}

func (sas Service) GetScheduledArticle(ctx context.Context, articleID string) (*ScheduledArticle, error) {
	start := time.Now()
	dart, err := sas.Querier.GetArticle(ctx, nullString(articleID))
	sas.Logger.Printf("queried GetArticle in %v", time.Since(start))
	if err != nil {
		return nil, db.StandardizeErr(err)
	}
	var schArticle ScheduledArticle
	if err = schArticle.fromDB(dart); err != nil {
		return nil, err
	}
	return &schArticle, nil
}

func (sas Service) SaveScheduledArticle(ctx context.Context, article *ScheduledArticle) error {
	// TODO: Make less racey
	if article.ScheduleFor != nil &&
		article.ScheduleFor.Before(time.Now().Add(5*time.Minute)) {
		article.ScheduleFor = nil
		if err := article.Publish(ctx, sas.ContentStore); err != nil {
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
	*dart, err = sas.Querier.UpdateSpotlightPAArticle(ctx, db.UpdateSpotlightPAArticleParams{
		ArcID:           dart.ArcID,
		SpotlightPAData: dart.SpotlightPAData,
		ScheduleFor:     dart.ScheduleFor,
		SpotlightPAPath: dart.SpotlightPAPath,
	})
	sas.Logger.Printf("queried UpdateSpotlightPAArticle in %v", time.Since(start))
	if err != nil {
		err = db.StandardizeErr(err)
		return err
	}

	if err = article.fromDB(*dart); err != nil {
		return err
	}
	return nil
}

func (sas Service) PopScheduledArticles(ctx context.Context, callback func([]*ScheduledArticle) error) error {
	start := time.Now()
	poppedArts, err := sas.Querier.PopScheduled(ctx)
	sas.Logger.Printf("queried PopScheduled in %v", time.Since(start))
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

func (fs Service) GetArcStory(ctx context.Context, articleID string) (story *ArcStory, err error) {
	start := time.Now()
	dart, err := fs.Querier.GetArticle(ctx, nullString(articleID))
	fs.Printf("GetArticle query time: %v", time.Since(start))
	if err != nil {
		err = db.StandardizeErr(err)
		return
	}
	var newstory ArcStory
	if err = newstory.fromDB(&dart); err != nil {
		return
	}
	story = &newstory
	return

}

func (fs Service) GetAvailableFeed(ctx context.Context) (stories []ArcStory, err error) {
	start := time.Now()
	var dbArts []db.Article
	dbArts, err = fs.Querier.ListAvailableArticles(ctx)
	fs.Printf("ListAvailableArticles query time: %v", time.Since(start))
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

func (fs Service) SaveAlmanackArticle(ctx context.Context, article *ArcStory) error {
	start := time.Now()
	dart, err := fs.Querier.UpdateAlmanackArticle(ctx, db.UpdateAlmanackArticleParams{
		ArcID:  nullString(article.ID),
		Status: article.Status.dbstring(),
		Note:   article.Note,
	})
	fs.Printf("UpdateAlmanackArticle query time: %v", time.Since(start))
	if err != nil {
		err = db.StandardizeErr(err)
		return err
	}
	if err = article.fromDB(&dart); err != nil {
		return err
	}

	return nil
}

func (fs Service) StoreFeed(ctx context.Context, newfeed ArcAPI, update bool) (err error) {
	arcItems, err := json.Marshal(&newfeed.Contents)
	if err != nil {
		return err
	}
	start := time.Now()
	dbarts, err := fs.Querier.UpdateArcArticles(ctx, arcItems)
	fs.Printf("StoreFeed query time: %v", time.Since(start))
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
