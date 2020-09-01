package mailchimp

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/spotlightpa/almanack/internal/httpjson"
	"github.com/spotlightpa/almanack/pkg/common"
)

func FlagVar(fs *flag.FlagSet) func(c *http.Client) V3 {
	apiKey := fs.String("mcnl-api-key", "", "`API key` for MailChimp newsletter archive")
	listID := fs.String("mcnl-list-id", "", "List `ID` for MailChimp newsletter archive")

	return func(c *http.Client) V3 {
		return NewV3(*apiKey, *listID, c)
	}
}

func V3Client(apiKey string, c *http.Client) *http.Client {
	if c == nil {
		c = http.DefaultClient
	}
	newClient := new(http.Client)
	*newClient = *c
	newClient.Transport = V3Transport(apiKey, newClient.Transport)
	return newClient
}

func V3Transport(apiKey string, rt http.RoundTripper) http.RoundTripper {
	return rtFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.HasSuffix(r.URL.Host, "api.mailchimp.com") {
			return nil, fmt.Errorf("bad URL for MailChimp API: %v", r.URL)
		}
		if rt == nil {
			rt = http.DefaultTransport
		}
		newReq := r.Clone(r.Context())
		newReq.SetBasicAuth("", apiKey)
		return rt.RoundTrip(newReq)
	})
}

type rtFunc func(r *http.Request) (*http.Response, error)

func (rt rtFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}

type V3 struct {
	cl               *http.Client
	listCampaignsURL string
}

func mustURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

var baseListCampaigns = mustURL("https://dc.api.mailchimp.com/3.0/campaigns?count=10&offset=0&status=sent&fields=campaigns.archive_url,campaigns.send_time,campaigns.settings.subject_line,campaigns.settings.title&sort_field=send_time&sort_dir=desc")

func NewV3(apiKey, listID string, c *http.Client) V3 {
	cl := V3Client(apiKey, c)
	// API keys end with 123XYZ-us1, where us1 is the datacenter
	var datacenter string
	if n := strings.LastIndex(apiKey, "-"); n != -1 {
		datacenter = apiKey[n+1:]
	}
	u := new(url.URL)
	*u = *baseListCampaigns
	u.Host = fmt.Sprintf("%s.api.mailchimp.com", datacenter)
	q := u.Query()
	q.Set("list_id", listID)
	u.RawQuery = q.Encode()
	return V3{
		cl:               cl,
		listCampaignsURL: u.String(),
	}
}

func (v3 V3) ListCampaigns(ctx context.Context) (*ListCampaignsResp, error) {
	var data ListCampaignsResp
	if err := httpjson.Get(ctx, v3.cl, v3.listCampaignsURL, &data); err != nil {
		return nil, err
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
		Subject string `json:"subject_line"`
		Title   string `json:"title"`
	} `json:"settings"`
}

func (v3 V3) ListNewletters(ctx context.Context, kind string) ([]common.Newsletter, error) {
	resp, err := v3.ListCampaigns(ctx)
	if err != nil {
		return nil, err
	}
	newsletters := make([]common.Newsletter, 0, len(resp.Campaigns))
	for _, camp := range resp.Campaigns {
		if strings.HasPrefix(camp.Settings.Title, kind) {
			newsletters = append(newsletters, common.Newsletter{
				Subject:     camp.Settings.Subject,
				ArchiveURL:  camp.ArchiveURL,
				PublishedAt: camp.SentAt,
			})
		}
	}
	return newsletters, nil
}
