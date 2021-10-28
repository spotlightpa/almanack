package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/carlmjohnson/slackhook"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type Client struct {
	c *slackhook.Client
	l Logger
}

type (
	Message    = slackhook.Message
	Attachment = slackhook.Attachment
	Field      = slackhook.Field
)

func New(hookURL string, l Logger) Client {
	return Client{slackhook.New(hookURL, nil), l}
}

func (sc Client) Post(ctx context.Context, msg Message) error {
	if sc.c == nil {
		sc.printf("no slack client; skipping posting message")
		b, _ := json.MarshalIndent(&msg, "", "  ")
		fmt.Fprintf(os.Stderr, "%s\n", b)
		return nil
	}
	sc.printf("posting Slack message")
	return sc.c.PostCtx(ctx, msg)
}

func (sc Client) printf(format string, args ...interface{}) {
	if sc.l != nil {
		sc.l.Printf(format, args...)
	}
}
