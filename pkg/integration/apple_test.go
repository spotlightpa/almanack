package db_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/jsonfeed"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestPublishAppleNews(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	q := db.New(p)
	ctx := t.Context()
	cl := &http.Client{
		Transport: reqtest.Caching(almlog.HTTPTransport, "testdata/anf"),
	}
	http.DefaultClient.Transport = requests.ErrorTransport(errors.New("used default client"))
	svc := almanack.Services{
		Client:  cl,
		Queries: q,
		NewsFeed: &jsonfeed.NewsFeed{
			URL: "https://www.spotlightpa.org/feeds/full.json",
		},
	}
	be.NilErr(t, svc.UpdateAppleNewsArchive(ctx))
	newItems, err := svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.EqualLength(t, 15, newItems)

	be.NilErr(t, svc.PublishAppleNewsFeed(ctx))
	// Publishing should mark everyone as uploaded
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.Zero(t, newItems)

	// Updating archive should not mark previously uploaded items as null
	be.NilErr(t, svc.UpdateAppleNewsArchive(ctx))
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.EqualLength(t, 0, newItems)
}
