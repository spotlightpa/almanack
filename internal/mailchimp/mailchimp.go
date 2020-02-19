package mailchimp

import (
	"fmt"

	"github.com/mattbaird/gochimp"

	"github.com/spotlightpa/almanack/pkg/almanack"
)

func NewMailService(apiKey, listID string, l almanack.Logger) almanack.EmailService {
	if apiKey == "" || listID == "" {
		l.Printf("using mock mail service")
		return Mock{l}
	}
	return MailChimp{apiKey, listID, l}
}

type MailChimp struct {
	apiKey, listID string
	l              almanack.Logger
}

func (mc MailChimp) SendEmail(subject, body string) error {
	// Using MC APIv2 because in v3 they decided REST means
	// not being able to create and send a campign in any efficient way
	chimp := gochimp.NewChimp(mc.apiKey, true)
	resp, err := chimp.CampaignCreate(gochimp.CampaignCreate{
		Type: "plaintext",
		Options: gochimp.CampaignCreateOptions{
			Subject:   subject,
			ListID:    mc.listID,
			FromEmail: "press@spotlightpa.org",
			FromName:  "Spotlight PA",
		},
		Content: gochimp.CampaignCreateContent{
			Text: body,
		},
	})
	if err != nil {
		return err
	}
	mc.l.Printf("created campaign %q", resp.Id)
	resp2, err := chimp.CampaignSend(resp.Id)
	mc.l.Printf("sent %v", resp2.Complete)
	return err
}

type Mock struct {
	l almanack.Logger
}

func (mock Mock) SendEmail(subject, body string) error {
	mock.l.Printf("no MailChimp client, debugging output\n")
	fmt.Println(subject)
	fmt.Println("----")
	fmt.Println(body)
	return nil
}
