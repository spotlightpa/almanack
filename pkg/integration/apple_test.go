package db_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/feed2anf"
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
		ANF: &feed2anf.Service{
			NewsFeedURL: "https://www.spotlightpa.org/feeds/full.json",
		},
	}
	be.NilErr(t, svc.PublishAppleNewsFeed(ctx))
}
