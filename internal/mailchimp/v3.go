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

func AddFlags(fs *flag.FlagSet) func(c *http.Client) V3 {
	apiKey := fs.String("mcnl-api-key", "", "`API key` for MailChimp newsletter archive")
	listID := fs.String("mcnl-list-id", "", "List `ID` for MailChimp newsletter archive")

	return func(c *http.Client) V3 {
		return NewV3(*apiKey, *listID, c)
	}
}

type V3 struct {
	apiKey string
	listID string
	cl     *http.Client
}

func NewV3(apiKey, listID string, cl *http.Client) V3 {
	return V3{apiKey, listID, cl}
}

func (v3 V3) config(rb *requests.Builder) {
	// API keys end with 123XYZ-us1, where us1 is the datacenter
	_, datacenter, _ := strings.Cut(v3.apiKey, "-")
	rb.
		BaseURL("https://dc.api.mailchimp.com/3.0/").
		Client(v3.cl).
		BasicAuth("", v3.apiKey).
		Hostf("%s.api.mailchimp.com", datacenter)
}

func (v3 V3) ListCampaigns(ctx context.Context) (*ListCampaignsResp, error) {
	var data ListCampaignsResp
	if err := requests.
		New(v3.config).
		Path("campaigns").
		Param("list_id", v3.listID).
		Param("count", "10").
		Param("offset", "0").
		Param("status", "sent").
		Param("fields", "campaigns.archive_url,campaigns.send_time,campaigns.settings.subject_line,campaigns.settings.title,campaigns.settings.preview_text").
		Param("sort_field", "send_time").
		Param("sort_dir", "desc").
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

func (resp *ListCampaignsResp) ToNewsletters(mcKind string) []Newsletter {
	newsletters := make([]Newsletter, 0, len(resp.Campaigns))
	for _, camp := range resp.Campaigns {
		// Hacky but probably the best method?
		if strings.Contains(camp.Settings.Title, mcKind) {
			newsletters = append(newsletters, Newsletter{
				Subject:     camp.Settings.Subject,
				Blurb:       camp.Settings.PreviewText,
				Description: camp.Settings.Title,
				ArchiveURL:  camp.ArchiveURL,
				PublishedAt: camp.SentAt,
			})
		}
	}
	return newsletters
}
