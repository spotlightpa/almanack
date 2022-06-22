package mailchimp

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/requests"
)

func FlagVar(fs *flag.FlagSet) func(c *http.Client) V3 {
	apiKey := fs.String("mcnl-api-key", "", "`API key` for MailChimp newsletter archive")
	listID := fs.String("mcnl-list-id", "", "List `ID` for MailChimp newsletter archive")

	return func(c *http.Client) V3 {
		return NewV3(*apiKey, *listID, c)
	}
}

type V3 struct {
	listCampaignBuilder *requests.Builder
}

func NewV3(apiKey, listID string, c *http.Client) V3 {
	// API keys end with 123XYZ-us1, where us1 is the datacenter
	_, datacenter, _ := strings.Cut(apiKey, "-")

	return V3{
		requests.URL("https://dc.api.mailchimp.com/3.0/campaigns?count=10&offset=0&status=sent&fields=campaigns.archive_url,campaigns.send_time,campaigns.settings.subject_line,campaigns.settings.title,campaigns.settings.preview_text&sort_field=send_time&sort_dir=desc").
			Client(c).
			BasicAuth("", apiKey).
			Hostf("%s.api.mailchimp.com", datacenter).
			Param("list_id", listID),
	}
}

func (v3 V3) listCampaigns(ctx context.Context) (*ListCampaignsResp, error) {
	var data ListCampaignsResp
	if err := v3.listCampaignBuilder.
		Clone().
		ToJSON(&data).
		Fetch(ctx); err != nil {
		return nil, fmt.Errorf("could not list MC campaigns: %w", err)
	}
	return &data, nil
}

type ListCampaignsResp struct {
	Campaigns []Campaign `json:"campaigns"`
}

type Campaign struct {
	ArchiveURL string    `json:"archive_url"`
	SentAt     time.Time `json:"send_time"`
	Settings   struct {
		Subject     string `json:"subject_line"`
		Title       string `json:"title"`
		PreviewText string `json:"preview_text"`
	} `json:"settings"`
}

type Newsletter struct {
	Subject     string    `json:"subject"`
	Blurb       string    `json:"blurb"`
	Description string    `json:"description"`
	ArchiveURL  string    `json:"archive_url"`
	PublishedAt time.Time `json:"published_at"`
}

func (v3 V3) ListNewletters(ctx context.Context, kind string) ([]Newsletter, error) {
	resp, err := v3.listCampaigns(ctx)
	if err != nil {
		return nil, err
	}
	newsletters := make([]Newsletter, 0, len(resp.Campaigns))
	for _, camp := range resp.Campaigns {
		// Hacky but probably the best method?
		if strings.Contains(camp.Settings.Title, kind) {
			newsletters = append(newsletters, Newsletter{
				Subject:     camp.Settings.Subject,
				Blurb:       camp.Settings.PreviewText,
				Description: camp.Settings.Title,
				ArchiveURL:  camp.ArchiveURL,
				PublishedAt: camp.SentAt,
			})
		}
	}
	return newsletters, nil
}
