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
	feedKey    = "almanack.feed"
	statusKey  = "almanack.feed-status"
	statusLock = "almanack.feed-status.lock"
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
	err = fs.PopulateStatuses(feed.Contents)
	return feed, err
}

func (fs FeedService) getStatusIDs() (map[string]Status, error) {
	ids := map[string]Status{}
	if err := fs.DataStore.Get(statusKey, &ids); err != nil &&
		!errors.Is(err, errutil.NotFound) {
		return ids, err
	}
	return ids, nil
}

func (fs FeedService) IsAvailable(articleID string) error {
	ids, err := fs.getStatusIDs()
	if err != nil {
		return err
	}
	if ids[articleID] == StatusAvailable {
		return nil
	}
	return errutil.NotFound
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

func (fs FeedService) SetStatus(articleid string, status Status) error {
	unlock, err := fs.DataStore.GetLock(statusLock)
	if err != nil {
		return err
	}
	defer unlock()

	ids, err := fs.getStatusIDs()
	if err != nil {
		return err
	}

	ids[articleid] = status
	prune(ids)
	if err = fs.DataStore.Set(statusKey, &ids); err != nil {
		return err
	}
	return nil
}

func prune(ids map[string]Status) {
	for k, v := range ids {
		if v == StatusUnset {
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

func (fs FeedService) PopulateStatuses(stories []Contents) (err error) {
	ids, err := fs.getStatusIDs()
	if err != nil {
		return err
	}
	for i := range stories {
		story := &stories[i]
		story.Status = ids[story.ID]
	}
	return nil
}
