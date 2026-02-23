package youtube_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/youtube"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestService(t *testing.T) {
	almlog.UseTestLogger(t)
	svc := youtube.Service{
		ChannelID: "abc123",
	}
	cl := &http.Client{
		Transport: reqtest.Replay("testdata"),
	}
	entries, err := svc.FetchFeed(t.Context(), cl)
	be.NilErr(t, err)
	be.Nonzero(t, entries)
	for _, entry := range entries {
		be.Nonzero(t, entry)
	}
}
