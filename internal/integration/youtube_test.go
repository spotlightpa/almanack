package integration_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"github.com/spotlightpa/almanack/internal/services/github"
	"github.com/spotlightpa/almanack/internal/services/youtube"
)

func TestYouTube(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	svc := almsvc.Services{
		DB:      dbhandle,
		Queries: dbhandle.Queries(),
		YT: &youtube.Feed{
			ChannelID: "abc123",
		},
		Client: &http.Client{
			Transport: requests.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				// Intercept img.youtube.com from youtube.BestThumbnailURL and subsequent GET
				if req.URL.Host == "img.youtube.com" {
					return reqtest.ReplayFile("testdata/youtube/NCmg292P.res.txt").RoundTrip(req)
				}
				return reqtest.Replay("testdata/youtube").RoundTrip(req)
			}),
		},
		FileStore:    aws.NewTestBlobStore(t.ArtifactDir(), "file"),
		ImageStore:   aws.NewTestBlobStore(t.ArtifactDir(), "image"),
		ContentStore: github.NewMockClient(t.ArtifactDir(), "github"),
	}
	ctx := t.Context()
	{ // Should not have pages
		pages := be.OK(svc.Queries.ListPages(ctx, db.ListPagesParams{
			FilePath: "content/videos/%",
			Limit:    20,
			Offset:   0,
		}))
		be.Falsey(pages)
	}
	{ // Load initial items
		be.Falsey(svc.UpdateYouTubeFeed(ctx))
	}
	{ // Should have pages
		pages := be.OK(svc.Queries.ListPages(ctx, db.ListPagesParams{
			FilePath: "content/videos/%",
			Limit:    20,
			Offset:   0,
		}))
		be.EqualLength(pages, 15)
	}
}
