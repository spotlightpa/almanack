package integration_test

import (
	"errors"
	"net/http"
	"strings"
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
	// Return an anf.Response unless it's the section lookup
	var anfRT requests.RoundTripFunc = func(req *http.Request) (*http.Response, error) {
		var res any = anf.Response{
			Data: anf.ResponseData{
				ID: "abc123",
			},
		}
		if strings.Contains(req.URL.Path, "/sections") {
			res = anf.ListSectionResponse{}
		}
		return reqtest.ReplayJSON(200, &res).RoundTrip(req)
	}
	svc := almanack.Services{
		Client:  cl,
		Queries: q,
		NewsFeed: &jsonfeed.NewsFeed{
			URL: "https://www.spotlightpa.org/feeds/full.json",
		},
		ANF: &anf.Service{Client: &http.Client{
			Transport: anfRT,
		}},
	}

	// Updating archive should add unuploaded items
	be.NilErr(t, svc.NewsFeed.UpdateAppleNewsArchive(ctx, svc.Client, svc.Queries))
	newItems, err := svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.EqualLength(t, 15, newItems)

	// Publishing should mark everything as uploaded
	be.NilErr(t, svc.PublishAppleNewsFeed(ctx))
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.Zero(t, newItems)

	// Updating archive should not mark previously uploaded items as null
	be.NilErr(t, svc.NewsFeed.UpdateAppleNewsArchive(ctx, svc.Client, svc.Queries))
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx)
	be.NilErr(t, err)
	be.EqualLength(t, 0, newItems)
}
