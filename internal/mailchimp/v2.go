package mailchimp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/mattbaird/gochimp"

	"github.com/spotlightpa/almanack/pkg/common"
)

type EmailService interface {
	SendEmail(ctx context.Context, subject, body string) error
}

func NewMailService(apiKey, listID string, c *http.Client) EmailService {
	if apiKey == "" || listID == "" {
		common.Logger.Printf("using mock mail service")
		return MockEmailService{}
	}
	return V2{apiKey, listID, c}
}

// V2 uses MC APIv2 because in v3 they decided REST means
// not being able to create and send a campign in any efficient way
type V2 struct {
	apiKey, listID string
	c              *http.Client
}

func (v2 V2) SendEmail(ctx context.Context, subject, body string) (err error) {
	defer func() {
		if err != nil {
			err = resperr.New(http.StatusBadGateway, "MailChimp problem: %w", err)
		}
	}()
	defer errutil.Trace(&err)

	// API keys end with 123XYZ-us1, where us1 is the datacenter
	_, datacenter, _ := strings.Cut(v2.apiKey, "-")
	var resp gochimp.CampaignResponse
	err = requests.
		URL("https://test.api.mailchimp.com/2.0/campaigns/create.json").
		Hostf("%s.api.mailchimp.com", datacenter).
		Client(v2.c).
		BodyJSON(gochimp.CampaignCreate{
			ApiKey: v2.apiKey,
			Type:   "plaintext",
			Options: gochimp.CampaignCreateOptions{
				Subject:   subject,
				ListID:    v2.listID,
				FromEmail: "press@spotlightpa.org",
				FromName:  "Spotlight PA",
			},
			Content: gochimp.CampaignCreateContent{
				Text: body,
			},
		}).
		ToJSON(&resp).
		Fetch(ctx)
	if err != nil {
		return err
	}
	common.Logger.Printf("created campaign %q", resp.Id)
	type v2CampaignSend struct {
		APIKey     string `json:"apikey"`
		CampaignID string `json:"cid"`
	}
	var resp2 gochimp.CampaignSendResponse
	err = requests.
		URL("https://test.api.mailchimp.com/2.0/campaigns/send.json").
		Hostf("%s.api.mailchimp.com", datacenter).
		Client(v2.c).
		BodyJSON(v2CampaignSend{
			APIKey:     v2.apiKey,
			CampaignID: resp.Id,
		}).
		ToJSON(&resp2).
		Fetch(ctx)
	common.Logger.Printf("sent %v", resp2.Complete)
	return err
}

type MockEmailService struct {
}

func (mock MockEmailService) SendEmail(ctx context.Context, subject, body string) error {
	common.Logger.Printf("no MailChimp client, debugging output\n")
	fmt.Println(subject)
	fmt.Println("----")
	fmt.Println(body)
	return nil
}
