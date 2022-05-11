package almanack

import (
	"context"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/jackc/pgtype"
	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/db"
)

func (svc Service) GetArcStory(ctx context.Context, articleID string) (story *ArcStory, err error) {
	defer errutil.Trace(&err)

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
	defer errutil.Trace(&err)

	pager := db.PageNumSize(page, 20)
	start := time.Now()
	dbArts, err := db.Paginate(
		pager, ctx,
		svc.Queries.ListAvailableArticles,
		db.ListAvailableArticlesParams{
			Offset: pager.Offset(),
			Limit:  pager.Limit(),
		})
	svc.Printf("ListAvailableArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}

	nextPage = pager.NextPage

	stories, err = storiesFromDB(dbArts)
	return
}

func (svc Service) SaveAlmanackArticle(ctx context.Context, article *ArcStory, setArcData bool) (err error) {
	defer errutil.Trace(&err)

	arcData := db.NullJSONB
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
	defer errutil.Trace(&err)

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
	defer errutil.Trace(&err)

	pager := db.PageNumSize(page, 50)
	start := time.Now()
	dbArts, err := db.Paginate(
		pager, ctx,
		svc.Queries.ListAllArcArticles,
		db.ListAllArcArticlesParams{
			Limit:  pager.Limit(),
			Offset: pager.Offset(),
		})
	svc.Printf("ListAllArticles query time: %v", time.Since(start))
	if err != nil {
		return
	}
	nextPage = pager.NextPage
	stories, err = storiesFromDB(dbArts)
	return
}
