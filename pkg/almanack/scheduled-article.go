package almanack

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

type ScheduledArticle struct {
	Article
	ScheduleFor *time.Time
	LastArcSync time.Time
	LastSaved   *time.Time
}

func (schArticle *ScheduledArticle) ResetArcData(dbArticle db.Article) error {
	schArticle.LastArcSync = time.Now()
	var arcStory ArcStory
	if err := json.Unmarshal(dbArticle.ArcData, &arcStory); err != nil {
		return err
	}
	art, err := arcStory.ToArticle()
	if err != nil {
		return err
	}
	schArticle.Article = *art
	return nil
}

type ScheduledArticleService struct {
	ArticleService
	DataStore
	ContentStore
	Logger
	Querier db.Querier
}

func (sas ScheduledArticleService) get(articleID string) (*ScheduledArticle, error) {
	var data ScheduledArticle
	err := sas.DataStore.Get("almanack.scheduled-article."+articleID, &data)
	switch {
	case err == nil:
		return &data, nil
	default:
		return nil, err
	case errors.Is(err, errutil.NotFound):
		return nil, nil
	}
}

func (sas ScheduledArticleService) set(article *ScheduledArticle) error {
	articleID := article.ArcID
	return sas.DataStore.Set("almanack.scheduled-article."+articleID, article)
}

func (sas ScheduledArticleService) lock() (unlock func(), err error) {
	return sas.DataStore.GetLock("almanack.scheduled-articles-lock")
}

// Get the existing list of scheduled articles
func (sas ScheduledArticleService) listIDs() (map[string]bool, error) {
	ids := map[string]bool{}
	err := sas.DataStore.Get("almanack.scheduled-articles-list", &ids)
	if errors.Is(err, errutil.NotFound) {
		err = nil
	}
	prune(ids)
	return ids, err
}

func (sas ScheduledArticleService) setIDs(ids map[string]bool) error {
	prune(ids)
	return sas.DataStore.Set("almanack.scheduled-articles-list", &ids)
}

func prune(ids map[string]bool) {
	for k, v := range ids {
		if !v {
			delete(ids, k)
		}
	}
}

func (sas ScheduledArticleService) Get(ctx context.Context, articleID string) (*ScheduledArticle, error) {
	start := time.Now()
	dart, err := sas.Querier.GetArticle(ctx, nullString(articleID))
	sas.Logger.Printf("queried GetArticle in %v", time.Since(start))
	if err != nil {
		return nil, db.StandardizeErr(err)
	}
	return ScheduledArticleFromDB(dart)
}

func (sas ScheduledArticleService) Save(ctx context.Context, article *ScheduledArticle) error {
	// Get the lock
	unlock, err := sas.lock()
	if err != nil {
		return err
	}
	defer unlock()

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
	if err := sas.set(article); err != nil {
		return err
	}

	ids, err := sas.listIDs()
	if err != nil {
		return err
	}

	// If the status of the article changed, update the list
	shouldPub := article.ScheduleFor != nil
	hasChanged := shouldPub != ids[article.ArcID]

	if hasChanged {
		ids[article.ArcID] = shouldPub
		if err := sas.setIDs(ids); err != nil {
			return err
		}
	}

	return nil
}

func (sas ScheduledArticleService) PopScheduledArticles(ctx context.Context, callback func([]*ScheduledArticle) error) error {
	start := time.Now()
	poppedArts, err := sas.Querier.PopScheduled(ctx)
	sas.Logger.Printf("queried PopScheduled in %v", time.Since(start))
	if err != nil {
		return err
	}
	overdueArts := make([]*ScheduledArticle, len(poppedArts))
	for i := range overdueArts {
		overdueArts[i], err = ScheduledArticleFromDB(poppedArts[i])
		if err != nil {
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

func ScheduledArticleFromDB(dbArticle db.Article) (*ScheduledArticle, error) {
	var schArticle ScheduledArticle
	schArticle.ArcID = dbArticle.ArcID.String
	if dbArticle.ScheduleFor.Valid {
		t := dbArticle.ScheduleFor.Time
		schArticle.ScheduleFor = &t
	}
	spotlightPADataVal := map[string]interface{}{
		"ID":               &schArticle.ID,
		"Budget":           &schArticle.Budget,
		"ImageURL":         &schArticle.ImageURL,
		"ImageCaption":     &schArticle.ImageCaption,
		"ImageCredit":      &schArticle.ImageCredit,
		"PubDate":          &schArticle.PubDate,
		"Slug":             &schArticle.Slug,
		"Authors":          &schArticle.Authors,
		"Byline":           &schArticle.Byline,
		"Hed":              &schArticle.Hed,
		"Subhead":          &schArticle.Subhead,
		"Summary":          &schArticle.Summary,
		"Blurb":            &schArticle.Blurb,
		"Kicker":           &schArticle.Kicker,
		"Body":             &schArticle.Body,
		"LinkTitle":        &schArticle.LinkTitle,
		"SuppressFeatured": &schArticle.SuppressFeatured,
		"LastArcSync":      &schArticle.LastArcSync,
		"LastSaved":        &schArticle.LastSaved,
	}
	if err := json.Unmarshal(dbArticle.SpotlightPAData, &spotlightPADataVal); err != nil {
		return nil, err
	}

	if schArticle.LastArcSync.IsZero() {
		if err := schArticle.ResetArcData(dbArticle); err != nil {
			return nil, err
		}
	}
	return &schArticle, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{}
}

func ScheduledArticleToDB(schArticle *ScheduledArticle) (*db.Article, error) {
	var dart db.Article
	if schArticle.ArcID != "" {
		dart.ArcID = nullString(schArticle.ArcID)
	}
	dart.ScheduleFor = nullTime(schArticle.ScheduleFor)
	spotlightPADataVal := map[string]interface{}{
		"ID":               &schArticle.ID,
		"Budget":           &schArticle.Budget,
		"ImageURL":         &schArticle.ImageURL,
		"ImageCaption":     &schArticle.ImageCaption,
		"ImageCredit":      &schArticle.ImageCredit,
		"PubDate":          &schArticle.PubDate,
		"Slug":             &schArticle.Slug,
		"Authors":          &schArticle.Authors,
		"Byline":           &schArticle.Byline,
		"Hed":              &schArticle.Hed,
		"Subhead":          &schArticle.Subhead,
		"Summary":          &schArticle.Summary,
		"Blurb":            &schArticle.Blurb,
		"Kicker":           &schArticle.Kicker,
		"Body":             &schArticle.Body,
		"LinkTitle":        &schArticle.LinkTitle,
		"SuppressFeatured": &schArticle.SuppressFeatured,
		"LastArcSync":      &schArticle.LastArcSync,
		"LastSaved":        &schArticle.LastSaved,
	}
	var err error
	if dart.SpotlightPAData, err = json.Marshal(&spotlightPADataVal); err != nil {
		return nil, err
	}
	return &dart, nil
}
