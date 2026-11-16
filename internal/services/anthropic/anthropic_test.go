package anthropic_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/anthropic"
)

func TestService(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	svc := anthropic.Service{
		Key: "abc123",
		Client: &http.Client{
			Transport: reqtest.Replay("testdata/"),
		},
	}
	res := be.OK(svc.CreateMessage(t.Context(), anthropic.CreateMessageRequest{
		Model:     "claude-sonnet-5-5",
		MaxTokens: 300,
		System:    "You write alt text for web images. Output only the alt text, one or two concise sentences.",
		Messages: []anthropic.Message{
			{
				Role: "user",
				Content: []anthropic.Content{
					{
						Type: "image",
						Source: anthropic.Source{
							Type: "url",
							URL:  "https://www.spotlightpa.org/imgproxy/insecure/rt:fill/w:1000/h:800/g:ce/el:1/q:75/MjAyNi8xMS8wMW5mLTQyejItY2g2dy1oeWRqLmpwZWc=.webp",
						},
					},
					{
						Type: "text",
						Text: "Write alt text for this image.",
					},
				},
			},
		},
	}))
	be.
		Truthy(res).
		AtLeastLength(res.Content, 1).
		Truthy(res.Content[0].Text)

}
