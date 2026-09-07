package youtube_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/youtube"
)

func TestService(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	svc := youtube.Feed{
		ChannelID: "abc123",
	}
	cl := &http.Client{
		Transport: reqtest.Replay("testdata"),
	}
	entries := be.OK(svc.FetchFeed(t.Context(), cl))
	be.Truthy(entries)
	for _, entry := range entries {
		be.Truthy(entry)
	}
}

func TestBestThumbnailURL(t *testing.T) {
	type testcase struct {
		n    int
		want string
	}
	assert.Run(t, map[string]testcase{
		"maxres available":  {0, "/maxresdefault.jpg$"},
		"sd available":      {1, "/sddefault.jpg$"},
		"nothing available": {10, "/default.jpg$"},
	}, func(be assert.TB, tc testcase) {
		n := 0
		cl := &http.Client{
			Transport: requests.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				n++
				if n > tc.n {
					return reqtest.ReplayString("HTTP/1.1 200 OK\r\n\r\n").RoundTrip(req)
				}
				return reqtest.ReplayString("HTTP/1.1 404 Not Found\r\n\r\n").RoundTrip(req)
			}),
		}
		got := youtube.BestThumbnailURL(be.Context(), cl, "abc")
		be.Match(got, tc.want)
	})
}
