package mailchimp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/mattbaird/gochimp"

	"github.com/spotlightpa/almanack/internal/stringutils"
	"github.com/spotlightpa/almanack/pkg/common"
)

func NewMailService(apiKey, listID string, l common.Logger, c *http.Client) common.EmailService {
	if apiKey == "" || listID == "" {
		l.Printf("using mock mail service")
		return MockEmailService{l}
	}
	return EmailService{apiKey, listID, l, c}
}

// EmailService uses MC APIv2 because in v3 they decided REST means
// not being able to create and send a campign in any efficient way
type EmailService struct {
	apiKey, listID string
	l              common.Logger
	c              *http.Client
}

func (mc EmailService) SendEmail(ctx context.Context, subject, body string) (err error) {
	defer func() {
		if err != nil {
			err = resperr.New(http.StatusBadGateway, "MailChimp problem: %w", err)
		}
	}()
	defer errutil.Trace(&err)

	// API keys end with 123XYZ-us1, where us1 is the datacenter
	_, datacenter, _ := stringutils.Cut(mc.apiKey, "-")
	var resp gochimp.CampaignResponse
	err = requests.
		URL("https://test.api.mailchimp.com/2.0/campaigns/create.json").
		Hostf("%s.api.mailchimp.com", datacenter).
		Client(mc.c).
		BodyJSON(gochimp.CampaignCreate{
			ApiKey: mc.apiKey,
			Type:   "plaintext",
			Options: gochimp.CampaignCreateOptions{
				Subject:   subject,
				ListID:    mc.listID,
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
	mc.l.Printf("created campaign %q", resp.Id)
	type v2CampaignSend struct {
		APIKey     string `json:"apikey"`
		CampaignID string `json:"cid"`
	}
	var resp2 gochimp.CampaignSendResponse
	err = requests.
		URL("https://test.api.mailchimp.com/2.0/campaigns/send.json").
		Hostf("%s.api.mailchimp.com", datacenter).
		Client(mc.c).
		BodyJSON(v2CampaignSend{
			APIKey:     mc.apiKey,
			CampaignID: resp.Id,
		}).
		ToJSON(&resp2).
		Fetch(ctx)
	mc.l.Printf("sent %v", resp2.Complete)
	return err
}

type MockEmailService struct {
	l common.Logger
}

func (mock MockEmailService) SendEmail(ctx context.Context, subject, body string) error {
	mock.l.Printf("no MailChimp client, debugging output\n")
	fmt.Println(subject)
	fmt.Println("----")
	fmt.Println(body)
	return nil
}
