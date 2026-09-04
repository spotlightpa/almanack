package google

import (
	"cmp"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
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
		gcl := be.OK(svc.GAClient(ctx))
		cl.Transport = reqtest.Record(gcl.Transport, "testdata")
	}
	pages := be.OK(svc.MostPopularNews(ctx, &cl))
	be.EqualLength(pages, 20)
}
