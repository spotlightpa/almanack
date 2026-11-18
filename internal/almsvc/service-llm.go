package almsvc

import (
	"context"
	"net/http"

	"github.com/earthboundkid/errorx/v2"
	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/almanack/internal/services/anthropic"
)

func (svc Services) DescribeImage(ctx context.Context, imgproxyPath string) (desc string, err error) {
	defer errorx.Trace(&err)

	const (
		model        = "claude-haiku-5-5"
		systemPrompt = `You write alt text for images on a Pennsylvannia based news website. Output only the alt text, one concise sentence in AP style.`
	)
	url, err := svc.ImageStore.SignGetURL(ctx, imgproxyPath)
	if err != nil {
		return "", resperr.E{S: http.StatusBadGateway, E: err}
	}
	res, err := svc.LLM.CreateMessage(ctx, anthropic.CreateMessageRequest{
		Model:     model,
		MaxTokens: 300,
		System:    systemPrompt,
		Messages: []anthropic.Message{{
			Role: "user",
			Content: []anthropic.Content{{
				Type:   "image",
				Source: anthropic.Source{Type: "url", URL: url},
			}}}}})
	if err != nil {
		return "", resperr.E{S: http.StatusBadGateway, E: err}
	}
	for _, content := range res.Content {
		if content.Type == "text" {
			return content.Text, nil
		}
	}
	return "", resperr.E{S: http.StatusBadGateway, M: "no text returned by LLM API"}
}
