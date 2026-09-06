package youtube_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/youtube"
)

func TestService(t *testing.T) {
	almlog.UseTestLogger(t)
	svc := youtube.Feed{
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

func TestBestThumbnailURL(t *testing.T) {
	for _, tc := range []struct {
		name    string
		succeed string // first quality suffix to return 200 for
	}{
		{"maxres available", "maxresdefault.jpg"},
		{"only hq available", "hqdefault.jpg"},
		{"only default available", "default.jpg"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cl := &http.Client{
				Transport: requests.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					if req.URL.Path == "/vi/abc/"+tc.succeed {
						return reqtest.ReplayString("HTTP/1.1 200 OK\r\n\r\n").RoundTrip(req)
					}
					return reqtest.ReplayString("HTTP/1.1 404 Not Found\r\n\r\n").RoundTrip(req)
				}),
			}
			got := youtube.BestThumbnailURL(t.Context(), cl, "abc")
			be.Match(t, tc.succeed+`$`, got)
		})
	}
}
