package integration_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/anf"
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
		Transport: reqtest.Replay("testdata/anf"),
	}
	http.DefaultClient.Transport = requests.ErrorTransport(errors.New("used default client"))
	svc := almanack.Services{
		Client:  cl,
		Queries: q,
	}
	nf := &jsonfeed.NewsFeed{
		URL: "https://www.spotlightpa.org/feeds/full.json",
	}
	anfsvc := &anf.Service{Client: &http.Client{
		Transport: reqtest.ReplayJSON(200, &anf.Response{
			Data: anf.Data{
				ID: "abc123",
			}})}}

	// Updating archive should add unuploaded items
	be.NilErr(t, nf.UpdateAppleNewsArchive(ctx, svc.Client, svc.Queries))
	newItems, err := svc.Queries.ListNewsFeedUpdates(ctx, nf.URL)
	be.NilErr(t, err)
	be.EqualLength(t, 15, newItems)

	// It shouldn't confuse the feeds
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx, "http://example.com")
	be.NilErr(t, err)
	be.Zero(t, newItems)

	// Publishing should mark everything as uploaded
	be.Zero(t, svc.PublishAppleNewsFeed(ctx, nf, anfsvc))
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx, nf.URL)
	be.NilErr(t, err)
	be.Zero(t, newItems)

	// Updating archive should not mark previously uploaded items as null
	be.NilErr(t, nf.UpdateAppleNewsArchive(ctx, svc.Client, svc.Queries))
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx, nf.URL)
	be.NilErr(t, err)
	be.Zero(t, newItems)
}
