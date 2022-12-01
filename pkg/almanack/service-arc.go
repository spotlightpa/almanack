package almanack

import (
	"context"
	"net/http"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/jackc/pgtype"
	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/paginate"
)

func (svc Services) FetchArcFeed(ctx context.Context) (*arc.API, error) {
	var feed arc.API
	// Timeout needs to leave enough time to report errors to Sentry before
	// AWS kills the Lambda…
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	if err := requests.
		URL(svc.arcFeedURL).
		Client(svc.Client).
		ToJSON(&feed).
		Fetch(ctx); err != nil {
		return nil, resperr.New(
			http.StatusBadGateway, "could not fetch Arc feed: %w", err)
	}
	return &feed, nil
}

func (svc Services) GetArcStory(ctx context.Context, articleID string) (story *ArcStory, err error) {
	defer errutil.Trace(&err)

	dart, err := svc.Queries.GetArticleByArcID(ctx, articleID)
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

func (svc Services) ListAvailableArcStories(ctx context.Context, page int32) (stories []ArcStory, nextPage int32, err error) {
	defer errutil.Trace(&err)

	pager := paginate.PageNumber(page)
	pager.PageSize = 20
	dbArts, err := paginate.List(
		pager, ctx,
		svc.Queries.ListAvailableArticles,
		db.ListAvailableArticlesParams{
			Offset: pager.Offset(),
			Limit:  pager.Limit(),
		})
	if err != nil {
		return
	}

	nextPage = pager.NextPage

	stories, err = storiesFromDB(dbArts)
	return
}

func (svc Services) SaveAlmanackArticle(ctx context.Context, article *ArcStory, setArcData bool) (err error) {
	defer errutil.Trace(&err)

	arcData := db.NullJSONB
	if setArcData {
		if err = arcData.Set(article.FeedItem); err != nil {
			return err
		}
	}
	dart, err := svc.Queries.UpdateAlmanackArticle(ctx, db.UpdateAlmanackArticleParams{
		ArcID:      article.ID,
		Status:     article.Status.dbstring(),
		Note:       article.Note,
		SetArcData: setArcData,
		ArcData:    arcData,
	})
	if err != nil {
		return err
	}
	if err = article.fromDB(&dart); err != nil {
		return err
	}

	return nil
}

func (svc Services) StoreFeed(ctx context.Context, newfeed *arc.API) (err error) {
	defer errutil.Trace(&err)

	var arcItems pgtype.JSONB
	if err := arcItems.Set(&newfeed.Contents); err != nil {
		return err
	}
	return svc.Queries.UpdateArcArticles(ctx, arcItems)
}

func (svc Services) ListAllArcStories(ctx context.Context, page int32) (stories []ArcStory, nextPage int32, err error) {
	defer errutil.Trace(&err)

	pager := paginate.PageNumber(page)
	pager.PageSize = 50
	dbArts, err := paginate.List(
		pager, ctx,
		svc.Queries.ListAllArcArticles,
		db.ListAllArcArticlesParams{
			Limit:  pager.Limit(),
			Offset: pager.Offset(),
		})
	if err != nil {
		return
	}
	nextPage = pager.NextPage
	stories, err = storiesFromDB(dbArts)
	return
}
