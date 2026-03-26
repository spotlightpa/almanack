package integration_test

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/youtube"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestYouTube(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	svc := almanack.Services{
		Queries: db.New(p),
		Tx:      db.NewTxable(p),
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
	{ // Should have uploaded feeds/youtube-shorts.json
		feedfile := filepath.Join(t.ArtifactDir(), "file/feeds/youtube-shorts.json")
		var data struct {
			Videos []youtube.FeedItem `json:"videos"`
		}
		testfile.ReadJSON(t, feedfile, &data)
		be.EqualLength(t, 8, data.Videos)
		be.Nonzero(t, data.Videos[0].Description)
	}
	{ // Should have uploaded feeds/youtube-regular.json
		feedfile := filepath.Join(t.ArtifactDir(), "file/feeds/youtube-regular.json")
		var data struct {
			Videos []youtube.FeedItem `json:"videos"`
		}
		testfile.ReadJSON(t, feedfile, &data)
		be.EqualLength(t, 7, data.Videos)
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
