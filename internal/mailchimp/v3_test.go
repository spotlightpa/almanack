package mailchimp_test

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

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
	if err != nil {
		t.Fatalf("err != nil: %v", err)
	}
	if len(res.Campaigns) == 0 {
		t.Fatal("no campaigns found")
	}
	for _, c := range res.Campaigns {
		if _, err := url.Parse(c.ArchiveURL); c.ArchiveURL == "" || err != nil {
			t.Errorf("received bad archive URL: %q", c.ArchiveURL)
		}
		if c.SentAt.IsZero() {
			t.Errorf("missing send time for %q", c.ArchiveURL)
		}
		if c.Settings.Subject == "" {
			t.Errorf("missing subject line for %q", c.ArchiveURL)
		}
		if c.Settings.Title == "" {
			t.Errorf("missing title for %q", c.ArchiveURL)
		}
		if c.Settings.PreviewText == "" {
			t.Errorf("missing preview_text for %q", c.ArchiveURL)
		}
	}
	camp := res.Campaigns[0]
	body, err := mailchimp.ImportPage(context.Background(), &cl, camp.ArchiveURL)
	if err != nil {
		t.Fatalf("problem getting campaign page: %v", err)
	}
	expect, err := os.ReadFile("testdata/body.html")
	if err != nil {
		t.Fatalf("problem reading campaign example: %v", err)
	}
	if string(expect) != body {
		os.WriteFile("testdata/got.html", []byte(body), 0644)
		t.Fatal("unexpected body")
	}
}
