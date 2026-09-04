package google

import (
	"cmp"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"net/http"
	"os"
	"testing"
)

func TestMostPopularNews(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	svc := Service{
		viewID: cmp.Or(os.Getenv("ALMANACK_GOOGLE_TEST_VIEW"), "1")}
	ctx := t.Context()
	cl := *http.DefaultClient
	cl.Transport = reqtest.Replay("testdata")
	if os.Getenv("ALMANACK_GOOGLE_TEST_RECORD_REQUEST") != "" {
		gcl, err := svc.GAClient(ctx)
		be.Zero(err)
		cl.Transport = reqtest.Record(gcl.Transport, "testdata")
	}
	pages, err := svc.MostPopularNews(ctx, &cl)
	be.Zero(err)
	be.EqualLength(pages, 20)
}
