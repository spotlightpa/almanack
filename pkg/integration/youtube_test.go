package integration_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/youtube"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestYouTube(t *testing.T) {
	almlog.UseTestLogger(t)
	p := createTestDB(t)
	q := db.New(p)
	svc := almanack.Services{
		Queries: q,
		YT: &youtube.Feed{
			ChannelID: "abc123",
		},
		Client: &http.Client{
			Transport: reqtest.Replay("testdata/youtube"),
		},
	}
	ctx := t.Context()
	{ // Nothing in table initially
		items, err := q.ListYouTubeUpdates(ctx)
		be.NilErr(t, err)
		be.Zero(t, items)
	}
	{ // Load initial items
		be.NilErr(t, svc.UpdateYouTubeFeed(ctx))
	}
	var nItems int
	var someitem *db.Youtube
	{ // Should have items loaded
		items, err := q.ListYouTubeUpdates(ctx)
		be.NilErr(t, err)
		be.Nonzero(t, items)

		nItems = len(items)
		someitem = &items[0]
	}
	{ // Shouldn't get new items from refetching
		be.NilErr(t, svc.UpdateYouTubeFeed(ctx))
		items, err := q.ListYouTubeUpdates(ctx)
		be.NilErr(t, err)
		be.Nonzero(t, items)
		be.Equal(t, nItems, len(items))
	}
	{ // Set one to 'uploaded'
		item, err := q.UpdateYouTubeUploaded(ctx, someitem.ID)
		be.NilErr(t, err)
		be.Equal(t, someitem.ID, item.ID)
		// One less item remains
		nItems--
	}
	{ // Should still have the right number of items
		be.NilErr(t, svc.UpdateYouTubeFeed(ctx))
		items, err := q.ListYouTubeUpdates(ctx)
		be.NilErr(t, err)
		be.Nonzero(t, items)
		be.Equal(t, nItems, len(items))
	}
}
