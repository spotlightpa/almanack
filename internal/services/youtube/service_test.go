package youtube_test

import (
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/youtube"
	"net/http"
	"testing"
)

func TestService(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	svc := youtube.Feed{
		ChannelID: "abc123",
	}
	cl := &http.Client{
		Transport: reqtest.Replay("testdata"),
	}
	entries := be.OK(svc.FetchFeed(t.Context(), cl))
	be.NotZero(entries)
	for _, entry := range entries {
		be.NotZero(entry)
	}
}
