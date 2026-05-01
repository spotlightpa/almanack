package integration_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"github.com/spotlightpa/almanack/internal/services/github"
	"github.com/spotlightpa/almanack/internal/services/youtube"
)

func TestYouTube(t *testing.T) {
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
		pages, err := svc.Queries.ListPages(ctx, db.ListPagesParams{
			FilePath: "content/videos/%",
			Limit:    20,
			Offset:   0,
		})
		be.NilErr(t, err)
		be.Zero(t, pages)
	}
	{ // Load initial items
		be.NilErr(t, svc.UpdateYouTubeFeed(ctx))
	}
	{ // Should have pages
		pages, err := svc.Queries.ListPages(ctx, db.ListPagesParams{
			FilePath: "content/videos/%",
			Limit:    20,
			Offset:   0,
		})
		be.NilErr(t, err)
		be.EqualLength(t, 15, pages)
	}
}
