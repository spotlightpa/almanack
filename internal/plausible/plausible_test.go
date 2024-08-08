package plausible_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/plausible"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestAPI(t *testing.T) {
	almlog.UseTestLogger(t)

	site := os.Getenv("TEST_PLAUSIBLE_SITE")
	token := os.Getenv("TEST_PLAUSIBLE_TOKEN")
	client := http.Client{
		Transport: reqtest.Replay("testdata"),
	}
	if os.Getenv("RECORD") != "" {
		client.Transport = reqtest.Caching(nil, "testdata")
	}
	api := plausible.API{site, token}
	pages, err := api.MostPopularNews(context.Background(), &client)
	be.NilErr(t, err)
	be.Nonzero(t, len(pages))
}
