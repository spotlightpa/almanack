package almanack

import (
	"errors"
	"time"

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
	Logger
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

func (sas ScheduledArticleService) Save(article *ScheduledArticle) error {
	// Get the lock
	unlock, err := sas.lock()
	if err != nil {
		return err
	}
	defer unlock()

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

func (sas ScheduledArticleService) PopScheduledArticles(callback func([]*ScheduledArticle) error) error {
	unlock, err := sas.lock()
	if err != nil {
		return err
	}
	defer unlock()

	// Get the existing list of scheduled articles
	ids, err := sas.listIDs()
	if err != nil {
		return err
	}

	overdueArts := make([]*ScheduledArticle, 0, len(ids))

	// Get the articles
	for articleID := range ids {
		article, err := sas.get(articleID)
		if err != nil {
			return err
		}
		if article == nil {
			// Weird, log it
			sas.Logger.Printf("got unexpected nil article for ID %s", articleID)
			continue
		}
		// If it's passed due, send to callback
		shouldPub := article.ScheduleFor != nil && article.ScheduleFor.Before(time.Now())
		if !shouldPub {
			continue
		}
		overdueArts = append(overdueArts, article)
		delete(ids, articleID)
	}

	// If the status of the article changed, fire callback then update the list
	if len(overdueArts) > 0 {
		if err := callback(overdueArts); err != nil {
			return err
		}
		return sas.setIDs(ids)
	}
	return nil
}
