package mailchimp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spotlightpa/almanack/pkg/almlog"
)

type EmailService interface {
	SendEmail(ctx context.Context, subject, body string) error
}

func NewMailService(apiKey, listID string, c *http.Client) EmailService {
	if apiKey == "" || listID == "" {
		almlog.Logger.Warn("mocking email service")
		return MockEmailService{}
	}
	return V3{apiKey, listID, c}
}

type MockEmailService struct {
}

func (mock MockEmailService) SendEmail(ctx context.Context, subject, body string) error {
	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "mocking campaign, debug output")
	fmt.Println()
	fmt.Println(subject)
	fmt.Println("----")
	fmt.Println(body)
	return nil
}
