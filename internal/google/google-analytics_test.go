package google

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/stringx"
)

func TestMostPopularNews(t *testing.T) {
	svc := Service{}
	svc.l = log.Default()
	svc.viewID = stringx.First(os.Getenv("ALMANACK_GOOGLE_TEST_VIEW"), "1")
	ctx := context.Background()
	cl := *http.DefaultClient
	cl.Transport = requests.Replay("testdata")
	if os.Getenv("ALMANACK_GOOGLE_TEST_RECORD_REQUEST") != "" {
		gcl, err := svc.GAClient(ctx)
		be.NilErr(t, err)
		cl.Transport = requests.Record(gcl.Transport, "testdata")
	}
	pages, err := svc.MostPopularNews(ctx, &cl)
	be.NilErr(t, err)
	be.True(t, len(pages) >= 20)
}
