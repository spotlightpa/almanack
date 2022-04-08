package mailchimp_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/mailchimp"
)

func TestV3(t *testing.T) {
	cl := *http.DefaultClient
	cl.Transport = requests.Replay("testdata")
	apiKey := os.Getenv("ALMANACK_MC_TEST_API_KEY")
	listID := os.Getenv("ALMANACK_MC_TEST_LISTID")

	if os.Getenv("ALMANACK_MC_TEST_RECORD_REQUEST") != "" {
		cl.Transport = requests.Record(nil, "testdata")
	}
	v3 := mailchimp.NewV3(apiKey, listID, &cl)
	res, err := v3.ListCampaigns(context.Background())
	be.NilErr(t, err)
	be.Nonzero(t, res.Campaigns)

	for _, c := range res.Campaigns {
		be.Nonzero(t, c.ArchiveURL)
		_, err := url.Parse(c.ArchiveURL)
		be.NilErr(t, err)
		be.Nonzero(t, c.ArchiveURL)
		be.Nonzero(t, c.SentAt)
		be.Nonzero(t, c.Settings.Subject)
		be.Nonzero(t, c.Settings.Title)
		be.Nonzero(t, c.Settings.PreviewText)
	}
	camp := res.Campaigns[0]
	body, err := mailchimp.ImportPage(context.Background(), &cl, camp.ArchiveURL)
	be.NilErr(t, err)
	expect, err := os.ReadFile("testdata/body.html")
	be.NilErr(t, err)
	be.Debug(t, func() {
		os.WriteFile("testdata/got.html", []byte(body), 0644)
	})
	be.Equal(t, string(expect), body)
}
