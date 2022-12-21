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
)

func (svc Services) RefreshArcFromFeed(ctx context.Context) (fatal bool, err error) {
	defer errutil.Trace(&err)

	feed, feedErr := svc.fetchArcFeed(ctx)
	if feedErr != nil {
		return false, feedErr
	}

	var arcItems pgtype.JSONB
	if err := arcItems.Set(&feed.Contents); err != nil {
		return true, err
	}

	if err := svc.Queries.UpdateArc(ctx, arcItems); err != nil {
		return true, err
	}

	return false, nil
}

func (svc Services) fetchArcFeed(ctx context.Context) (*arc.API, error) {
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
