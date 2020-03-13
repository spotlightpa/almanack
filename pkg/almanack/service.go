package almanack

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
)

type FeedService struct {
	DataStore
	Logger
	Querier db.Querier
}

type Status int8

const (
	StatusUnset Status = iota
	StatusPlanned
	StatusAvailable
)

var dbStatusToStatus = map[string]Status{
	"U": StatusUnset,
	"P": StatusPlanned,
	"A": StatusAvailable,
}

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

func (story *ArcStory) fromDB(dart *db.Article) error {
	if err := json.Unmarshal(dart.ArcData, story); err != nil {
		return err
	}
	var ok bool
	if story.Status, ok = dbStatusToStatus[dart.Status]; !ok {
		return errors.New("bad status flag in database")
	}
	return nil
}

func (fs FeedService) SaveSupplements(article *ArcStory) error {
	return nil
}

func (fs FeedService) StoreFeed(ctx context.Context, newfeed ArcAPI) (err error) {
	arcItems, err := json.Marshal(&newfeed.Contents)
	if err != nil {
		return err
	}
	start := time.Now()
	_, err = fs.Querier.UpdateArcArticles(ctx, arcItems)
	fs.Printf("arcjson.StoreFeed query time: %v", time.Since(start))
	return
}

func (fs FeedService) PopulateSuplements(stories []ArcStory) (err error) {
	return nil
}
