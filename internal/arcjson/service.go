package arcjson

import (
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

type FeedService struct {
	almanack.DataStore
	almanack.Logger
}

const feedKey = "almanack-worker.feed"

func (fs FeedService) GetFeed() (API, error) {
	var feed API
	err := fs.DataStore.Get(feedKey, &feed)
	return feed, err
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

// func (fs FeedService) UpdateFeed(newfeed API) (newcontents []Contents, err error) {
// 	var oldfeed API
// 	err = fs.DataStore.GetSet(feedKey, &oldfeed, &newfeed)
// 	if errors.Is(err, errutil.NotFound) {
// 		fs.Logger.Printf("warning: no old feed data")
// 		return nil, nil
// 	}
// 	if err != nil {
// 		return nil, err
// 	}

// 	// TODO: Better status checking
// 	newcontents = diffFeed(newfeed, oldfeed)
// 	return newcontents, nil
// }

func diffFeed(newfeed, oldfeed API) []Contents {
	readyids := make(map[string]bool, len(oldfeed.Contents))
	for _, story := range oldfeed.Contents {
		if story.Workflow.StatusCode >= StatusSlot {
			readyids[story.ID] = true
		}
	}
	var newstories []Contents
	for _, story := range newfeed.Contents {
		if story.Workflow.StatusCode >= StatusSlot &&
			!readyids[story.ID] {
			newstories = append(newstories, story)
		}
	}
	return newstories
}

func (fs FeedService) UpdateMailStatus(newstories []Contents) (filteredStories []Contents, err error) {
	sentStories := make([]bool, len(newstories))
	getters := make([]func() error, len(newstories))
	for i := range newstories {
		j := i // Fix closure value
		story := newstories[j]
		getters[j] = func() error {
			err := fs.DataStore.GetSet("almanack.sent-campaigns."+story.ID,
				&sentStories[j], true)
			if err == errutil.NotFound {
				return nil
			}
			return err
		}
	}
	if err = errutil.ExecParallel(getters...); err != nil {
		return nil, err
	}
	filteredStories = newstories[:0]
	for i, story := range newstories {
		if !sentStories[i] {
			filteredStories = append(filteredStories, story)
		}
	}
	return filteredStories, nil
}
