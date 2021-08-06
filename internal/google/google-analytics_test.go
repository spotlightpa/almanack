package google

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/stringutils"
)

func TestMostPopularNews(t *testing.T) {
	svc := Service{}
	svc.l = log.Default()
	svc.viewID = stringutils.First(os.Getenv("ALMANACK_GOOGLE_TEST_VIEW"), "1")
	ctx := context.Background()
	cl := *http.DefaultClient
	cl.Transport = requests.Replay("testdata")
	if os.Getenv("ALMANACK_GOOGLE_TEST_RECORD_REQUEST") != "" {
		gcl, err := svc.GAClient(ctx)
		if err != nil {
			t.Fatal(err)
		}
		cl.Transport = requests.Record(gcl.Transport, "testdata")
	}
	pages, err := svc.MostPopularNews(ctx, &cl)
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) < 20 {
		t.Fatalf("wrong number of pages: %d", len(pages))
	}
}
