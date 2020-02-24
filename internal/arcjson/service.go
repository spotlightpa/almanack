package arcjson

import (
	"errors"

	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

type FeedService struct {
	almanack.DataStore
	almanack.Logger
}

const (
	feedKey       = "almanack.feed"
	suplementKey  = "almanack.feed-suplements"
	suplementLock = "almanack.feed-suplements.lock"
)

type Status int8

const (
	StatusUnset Status = iota
	StatusPlanned
	StatusAvailable
)

func (fs FeedService) GetFeed() (API, error) {
	var feed API
	err := fs.DataStore.Get(feedKey, &feed)
	if err != nil {
		return feed, err
	}
	err = fs.PopulateSuplements(feed.Contents)
	return feed, err
}

type supplement struct {
	Statuses map[string]Status
	Notes    map[string]string
}

func (fs FeedService) getSuplements() (supplement, error) {
	sups := supplement{map[string]Status{}, map[string]string{}}
	if err := fs.DataStore.Get(suplementKey, &sups); err != nil &&
		!errors.Is(err, errutil.NotFound) {
		return sups, err
	}
	return sups, nil
}

func (fs FeedService) GetAvailableFeed() ([]Contents, error) {
	feed, err := fs.GetFeed()
	if err != nil {
		return nil, err
	}

	filteredContents := feed.Contents[:0]
	for _, item := range feed.Contents {
		if item.Status == StatusPlanned || item.Status == StatusAvailable {
			filteredContents = append(filteredContents, item)
		}
	}
	return filteredContents, nil
}

func (fs FeedService) SaveSupplements(article *Contents) error {
	articleID := article.ID
	unlock, err := fs.DataStore.GetLock(suplementLock)
	if err != nil {
		return err
	}
	defer unlock()

	sups, err := fs.getSuplements()
	if err != nil {
		return err
	}

	sups.Statuses[articleID] = article.Status
	sups.Notes[articleID] = article.Note
	pruneStatuses(sups.Statuses)
	pruneStr(sups.Notes)

	if err = fs.DataStore.Set(suplementKey, &sups); err != nil {
		return err
	}
	return nil
}

func pruneStatuses(ids map[string]Status) {
	for k, v := range ids {
		if v == StatusUnset {
			delete(ids, k)
		}
	}
}

func pruneStr(ids map[string]string) {
	for k, v := range ids {
		if v == "" {
			delete(ids, k)
		}
	}
}
func (fs FeedService) GetArticle(articleID string) (*almanack.Article, error) {
	feed, err := fs.GetFeed()
	if err != nil {
		return nil, err
	}
	content, err := feed.Get(articleID)
	if err != nil {
		return nil, err
	}
	article, err := content.ToArticle()
	if err != nil {
		return nil, err
	}
	return article, nil
}

func (fs FeedService) StoreFeed(newfeed API) (err error) {
	return fs.DataStore.Set(feedKey, &newfeed)
}

func (fs FeedService) PopulateSuplements(stories []Contents) (err error) {
	sups, err := fs.getSuplements()
	if err != nil {
		return err
	}
	for i := range stories {
		story := &stories[i]
		story.Status = sups.Statuses[story.ID]
		story.Note = sups.Notes[story.ID]
	}
	return nil
}
