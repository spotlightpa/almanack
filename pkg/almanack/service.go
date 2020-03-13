package almanack

import (
	"context"
	"encoding/json"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
)

type FeedService struct {
	DataStore
	Logger
	Querier db.Querier
}

// TODO
func (fs FeedService) GetArcStory(ctx context.Context, articleID string) (*ArcStory, error) {
	return nil, nil
}

func (fs FeedService) GetAvailableFeed(ctx context.Context) (stories []ArcStory, err error) {
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

// TODO
func (fs FeedService) SaveSupplements(article *ArcStory) error {
	return nil
}

func (fs FeedService) StoreFeed(ctx context.Context, newfeed ArcAPI, update bool) (err error) {
	arcItems, err := json.Marshal(&newfeed.Contents)
	if err != nil {
		return err
	}
	start := time.Now()
	dbarts, err := fs.Querier.UpdateArcArticles(ctx, arcItems)
	fs.Printf("arcjson.StoreFeed query time: %v", time.Since(start))
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

// TODO
func (fs FeedService) PopulateSuplements(stories []ArcStory) (err error) {
	return nil
}
