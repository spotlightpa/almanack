package almanack

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

type FeedService struct {
	DataStore
	Logger
	Querier db.Querier
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

var dbStatusToStatus = map[string]Status{
	"U": StatusUnset,
	"P": StatusPlanned,
	"A": StatusAvailable,
}

func (fs FeedService) GetFeed(ctx context.Context) (feed ArcAPI, err error) {
	start := time.Now()
	dbArts, err := fs.Querier.ListAllArticles(ctx)
	if err != nil {
		return
	}
	fs.Printf("ListAllArticles query time: %v", time.Since(start))
	feed.Contents = make([]ArcStory, len(dbArts))
	for i := range feed.Contents {
		if err = feed.Contents[i].fromDB(&dbArts[i]); err != nil {
			return
		}
	}
	return
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
func (fs FeedService) GetArticle(articleID string) (*Article, error) {
	feed, err := fs.GetFeed(context.Background())
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
