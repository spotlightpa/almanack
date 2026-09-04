package integration_test

import (
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"github.com/spotlightpa/almanack/internal/services/github"
	"github.com/spotlightpa/almanack/internal/services/youtube"
	"net/http"
	"testing"
)

func TestYouTube(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	svc := almsvc.Services{
		DB:      dbhandle,
		Queries: dbhandle.Queries(),
		YT: &youtube.Feed{
			ChannelID: "abc123",
		},
		Client: &http.Client{
			Transport: reqtest.Replay("testdata/youtube"),
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
		be.Zero(pages)
	}
	{ // Load initial items
		be.Zero(svc.UpdateYouTubeFeed(ctx))
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
