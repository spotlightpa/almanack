package almanack

import (
	"errors"
	"time"

	"github.com/spotlightpa/almanack/internal/errutil"
)

type ScheduledArticle struct {
	Article
	ScheduleFor *time.Time
	LastArcSync time.Time
}

type ScheduledArticleService struct {
	ArticleService
	DataStore
	Logger
}

func (sas ScheduledArticleService) Get(articleID string) (*ScheduledArticle, error) {
	var data ScheduledArticle
	err := sas.DataStore.Get("almanack.scheduled-article."+articleID, &data)
	switch {
	case err == nil:
		return &data, nil
	default:
		return nil, err
	case errors.Is(err, errutil.NotFound):
		// continue
	}

	sas.Logger.Printf("no article in datastore, falling back to article service")

	article, err := sas.ArticleService.GetArticle(articleID)
	if err != nil {
		return nil, err
	}
	data.Article = *article
	data.LastArcSync = time.Now()
	return &data, nil
}

func (sas ScheduledArticleService) Save(articleID string, article ScheduledArticle) error {
	// Get the lock
	unlock, err := sas.DataStore.GetLock("almanack.scheduled-articles-lock")
	defer unlock()
	if err != nil {
		return err
	}

	// Save the article
	if err := sas.DataStore.Set("almanack.scheduled-article."+articleID, &article); err != nil {
		return err
	}

	// Get the existing list of scheduled articles
	ids := map[string]bool{}
	if err = sas.DataStore.Get("almanack.scheduled-articles-list", &ids); err != nil &&
		!errors.Is(err, errutil.NotFound) {
		return err
	}

	// If the status of the article changed, update the list
	shouldPub := article.ScheduleFor != nil
	hasChanged := shouldPub != ids[articleID]

	if hasChanged {
		ids[articleID] = shouldPub
		if err := sas.DataStore.Set("almanack.scheduled-articles-list", &ids); err != nil {
			return err
		}
	}

	return nil
}
