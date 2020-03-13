package almanack

import (
	"context"
	"encoding/json"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
)

type FeedService struct {
	Logger
	Querier db.Querier
}

func (fs FeedService) GetArcStory(ctx context.Context, articleID string) (story *ArcStory, err error) {
	start := time.Now()
	dart, err := fs.Querier.GetArticle(ctx, nullString(articleID))
	fs.Printf("GetArticle query time: %v", time.Since(start))
	if err != nil {
		err = db.StandardizeErr(err)
		return
	}
	var newstory ArcStory
	if err = newstory.fromDB(&dart); err != nil {
		return
	}
	story = &newstory
	return

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

func (fs FeedService) SaveAlmanackArticle(ctx context.Context, article *ArcStory) error {
	start := time.Now()
	dart, err := fs.Querier.UpdateAlmanackArticle(ctx, db.UpdateAlmanackArticleParams{
		ArcID:  nullString(article.ID),
		Status: article.Status.dbstring(),
		Note:   article.Note,
	})
	fs.Printf("UpdateAlmanackArticle query time: %v", time.Since(start))
	if err != nil {
		err = db.StandardizeErr(err)
		return err
	}
	if err = article.fromDB(&dart); err != nil {
		return err
	}

	return nil
}

func (fs FeedService) StoreFeed(ctx context.Context, newfeed ArcAPI, update bool) (err error) {
	arcItems, err := json.Marshal(&newfeed.Contents)
	if err != nil {
		return err
	}
	start := time.Now()
	dbarts, err := fs.Querier.UpdateArcArticles(ctx, arcItems)
	fs.Printf("StoreFeed query time: %v", time.Since(start))
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
