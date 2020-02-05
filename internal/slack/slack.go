package slack

import (
	"github.com/carlmjohnson/slackhook"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

type Client struct {
	*slackhook.Client
	l almanack.Logger
}

type (
	Message    = slackhook.Message
	Attachment = slackhook.Attachment
	Field      = slackhook.Field
)

func New(hookURL string, l almanack.Logger) Client {
	return Client{slackhook.New(hookURL, nil), l}
}

func (sc Client) Post(msg Message) error {
	if sc.Client == nil {
		sc.l.Printf("no slack client; skipping posting message")
		return nil
	}
	sc.l.Printf("posting Slack message")
	return sc.Client.Post(msg)
}
