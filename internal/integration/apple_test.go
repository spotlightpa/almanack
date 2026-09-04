package integration_test

import (
	"errors"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/services/anf"
	"github.com/spotlightpa/almanack/internal/services/jsonfeed"
	"net/http"
	"strings"
	"testing"
)

func TestPublishAppleNews(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	q := dbhandle.Queries()
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
	svc := almsvc.Services{
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
	be.Zero(svc.NewsFeed.UpdateAppleNewsArchive(ctx, svc.Client, svc.Queries))
	newItems, err := svc.Queries.ListNewsFeedUpdates(ctx)
	be.Zero(err)
	be.EqualLength(newItems, 15)

	// Publishing should mark everything as uploaded
	be.Zero(svc.PublishAppleNewsFeed(ctx))
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx)
	be.Zero(err)
	be.Zero(newItems)

	// Updating archive should not mark previously uploaded items as null
	be.Zero(svc.NewsFeed.UpdateAppleNewsArchive(ctx, svc.Client, svc.Queries))
	newItems, err = svc.Queries.ListNewsFeedUpdates(ctx)
	be.Zero(err)
	be.EqualLength(newItems, 0)
}
