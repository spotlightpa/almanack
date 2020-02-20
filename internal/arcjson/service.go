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
	availableKey  = "almanack.feed-available"
	availableLock = "almanack.feed-available.lock"
)

func (fs FeedService) GetFeed() (API, error) {
	var feed API
	err := fs.DataStore.Get(feedKey, &feed)
	return feed, err
}

func (fs FeedService) getAvailableIDs() (map[string]bool, error) {
	ids := map[string]bool{}
	if err := fs.DataStore.Get(availableKey, &ids); err != nil &&
		!errors.Is(err, errutil.NotFound) {
		return ids, err
	}
	return ids, nil
}

func (fs FeedService) IsAvailable(articleID string) error {
	ids, err := fs.getAvailableIDs()
	if err != nil {
		return err
	}
	if ids[articleID] {
		return nil
	}
	return errutil.NotFound
}

func (fs FeedService) GetAvailableFeed() ([]Contents, error) {
	feed, err := fs.GetFeed()
	if err != nil {
		return nil, err
	}

	ids, err := fs.getAvailableIDs()
	if err != nil {
		return nil, err
	}
	filteredContents := feed.Contents[:0]
	for _, item := range feed.Contents {
		if ids[item.ID] {
			filteredContents = append(filteredContents, item)
		}
	}
	return filteredContents, nil
}

func (fs FeedService) SetAvailablity(articleid string, available bool) error {
	unlock, err := fs.DataStore.GetLock(availableLock)
	if err != nil {
		return err
	}
	defer unlock()

	ids, err := fs.getAvailableIDs()
	if err != nil {
		return err
	}

	ids[articleid] = available
	prune(ids)
	if err = fs.DataStore.Set(availableKey, &ids); err != nil {
		return err
	}
	return nil
}

func prune(ids map[string]bool) {
	for k, v := range ids {
		if !v {
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
