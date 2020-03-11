package almanack

import (
	"context"
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

func (sas ScheduledArticleService) Get(articleID string) (*ScheduledArticle, error) {
	data, err := sas.get(articleID)
	if data != nil || err != nil {
		return data, err
	}

	sas.Logger.Printf("no article in datastore, falling back to article service")

	article, err := sas.ArticleService.GetArticle(articleID)
	if err != nil {
		return nil, err
	}
	data = new(ScheduledArticle)
	data.Article = *article
	data.LastArcSync = time.Now()
	return data, nil
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
	poppedArts, err := sas.Querier.PopScheduled(ctx)
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

func ScheduledArticleFromDB(srcArticle db.Article) (*ScheduledArticle, error) {
	var sart ScheduledArticle
	sart.ArcID = srcArticle.ArcID.String
	if srcArticle.ScheduleFor.Valid {
		t := srcArticle.ScheduleFor.Time
		sart.ScheduleFor = &t
	}
	type spotlightPADataType struct {
		ID               *string
		Budget           *string
		ImageURL         *string
		ImageCaption     *string
		ImageCredit      *string
		PubDate          *time.Time
		Slug             *string
		Authors          *[]string
		Byline           *string
		Hed              *string
		Subhead          *string
		Summary          *string
		Blurb            *string
		Kicker           *string
		Body             *string
		LinkTitle        *string
		SuppressFeatured *bool
		LastArcSync      *time.Time
		LastSaved        **time.Time
	}
	spotlightPADataVal := spotlightPADataType{
		ID:               &sart.ID,
		Budget:           &sart.Budget,
		ImageURL:         &sart.ImageURL,
		ImageCaption:     &sart.ImageCaption,
		ImageCredit:      &sart.ImageCredit,
		PubDate:          &sart.PubDate,
		Slug:             &sart.Slug,
		Authors:          &sart.Authors,
		Byline:           &sart.Byline,
		Hed:              &sart.Hed,
		Subhead:          &sart.Subhead,
		Summary:          &sart.Summary,
		Blurb:            &sart.Blurb,
		Kicker:           &sart.Kicker,
		Body:             &sart.Body,
		LinkTitle:        &sart.LinkTitle,
		SuppressFeatured: &sart.SuppressFeatured,
		LastArcSync:      &sart.LastArcSync,
		LastSaved:        &sart.LastSaved,
	}
	if err := json.Unmarshal(srcArticle.SpotlightPAData, &spotlightPADataVal); err != nil {
		return nil, err
	}
	return &sart, nil
}
