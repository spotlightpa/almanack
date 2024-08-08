package google

import (
	"cmp"
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestMostPopularNews(t *testing.T) {
	almlog.UseTestLogger(t)
	svc := Service{}
	svc.viewID = cmp.Or(os.Getenv("ALMANACK_GOOGLE_TEST_VIEW"), "1")
	ctx := context.Background()
	cl := *http.DefaultClient
	cl.Transport = reqtest.Replay("testdata")
	if os.Getenv("ALMANACK_GOOGLE_TEST_RECORD_REQUEST") != "" {
		gcl, err := svc.GAClient(ctx)
		be.NilErr(t, err)
		cl.Transport = reqtest.Record(gcl.Transport, "testdata")
	}
	pages, err := svc.MostPopularNews(ctx, &cl)
	be.NilErr(t, err)
	be.True(t, len(pages) >= 20)
}
