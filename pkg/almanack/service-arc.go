package almanack

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/db"
)

func (svc Service) GetArcStory(ctx context.Context, articleID string) (story *ArcStory, err error) {
	start := time.Now()
	dart, err := svc.Queries.GetArticleByArcID(ctx, articleID)
	svc.Printf("GetArticleByArcID query time: %v", time.Since(start))
	if err != nil {
		err = db.NoRowsAs404(err, "could not find arc-id %q", articleID)
		return
	}
	var newstory ArcStory
	if err = newstory.fromDB(&dart); err != nil {
		return
	}
	story = &newstory
	return

}

func (svc Service) ListAvailableArcStories(ctx context.Context, page int32) (stories []ArcStory, nextPage int32, err error) {
	const pageSize = 20
	offset := page * pageSize

	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Queries.ListAvailableArticles(ctx, db.ListAvailableArticlesParams{
		Offset: offset,
		Limit:  pageSize + 1,
	})
	svc.Printf("ListAvailableArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}

	if len(dbArts) > pageSize {
		dbArts = dbArts[:pageSize]
		nextPage = page + 1
	}

	stories, err = storiesFromDB(dbArts)
	return
}

func (svc Service) SaveAlmanackArticle(ctx context.Context, article *ArcStory, setArcData bool) error {
	var (
		arcData pgtype.JSONB
		err     error
	)
	if setArcData {
		if err = arcData.Set(article.FeedItem); err != nil {
			return err
		}
	}
	start := time.Now()
	dart, err := svc.Queries.UpdateAlmanackArticle(ctx, db.UpdateAlmanackArticleParams{
		ArcID:      article.ID,
		Status:     article.Status.dbstring(),
		Note:       article.Note,
		SetArcData: setArcData,
		ArcData:    arcData,
	})
	svc.Printf("UpdateAlmanackArticle query time: %v", time.Since(start))
	if err != nil {
		return err
	}
	if err = article.fromDB(&dart); err != nil {
		return err
	}

	return nil
}

func (svc Service) StoreFeed(ctx context.Context, newfeed *arc.API) (err error) {
	var arcItems pgtype.JSONB
	if err := arcItems.Set(&newfeed.Contents); err != nil {
		return err
	}
	start := time.Now()
	err = svc.Queries.UpdateArcArticles(ctx, arcItems)
	svc.Printf("StoreFeed query time: %v", time.Since(start))
	return err
}

func (svc Service) ListAllArcStories(ctx context.Context, page int32) (stories []ArcStory, nextPage int32, err error) {
	const pageSize = 50
	offset := page * pageSize
	start := time.Now()
	var dbArts []db.Article
	dbArts, err = svc.Queries.ListAllArcArticles(ctx, db.ListAllArcArticlesParams{
		Limit:  pageSize + 1,
		Offset: offset,
	})
	svc.Printf("ListAllArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}
	if len(dbArts) > pageSize {
		dbArts = dbArts[:pageSize]
		nextPage = page + 1
	}
	stories, err = storiesFromDB(dbArts)
	return
}
