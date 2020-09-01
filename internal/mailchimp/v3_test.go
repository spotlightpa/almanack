package mailchimp_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/spotlightpa/almanack/internal/mailchimp"
)

func TestV3(t *testing.T) {
	apiKey := os.Getenv("ALMANACK_MC_TEST_API_KEY")
	listID := os.Getenv("ALMANACK_MC_TEST_LISTID")

	if apiKey == "" || listID == "" {
		t.Skip("Missing MailChimp ENV vars")
	}

	v3 := mailchimp.NewV3(apiKey, listID, nil)
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
	}
}
