package anthropic

import (
	"context"
	"errors"
	"flag"
	"net/http"

	"github.com/carlmjohnson/requests"
)

type Service struct {
	Key string
	*http.Client
}

func AddFlags(fl *flag.FlagSet) (svc *Service) {
	svc = new(Service)
	fl.StringVar(&svc.Key, "anthropic-api-key", "", "`key` for Anthropic API")
	svc.Client = &http.Client{
		Transport: requests.ErrorTransport(errors.New("anthropic client not configured")),
	}
	return svc
}

type CreateMessageRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []Message `json:"messages"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type   string `json:"type"`
	Source Source `json:"source,omitzero"`
	Text   string `json:"text,omitzero"`
}

type Source struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type CreateMessageResponse struct {
	Model        string    `json:"model"`
	ID           string    `json:"id"`
	Type         string    `json:"type"`
	Role         string    `json:"role"`
	Content      []Content `json:"content"`
	Container    any       `json:"container"`
	StopReason   string    `json:"stop_reason"`
	StopSequence any       `json:"stop_sequence"`
	StopDetails  any       `json:"stop_details"`
	Usage        Usage     `json:"usage"`
	Diagnostics  any       `json:"diagnostics"`
}

type CacheCreation struct {
	Ephemeral5MInputTokens int `json:"ephemeral_5m_input_tokens"`
	Ephemeral1HInputTokens int `json:"ephemeral_1h_input_tokens"`
}

type OutputTokensDetails struct {
	ThinkingTokens int `json:"thinking_tokens"`
}

type Usage struct {
	InputTokens              int                 `json:"input_tokens"`
	CacheCreationInputTokens int                 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int                 `json:"cache_read_input_tokens"`
	CacheCreation            CacheCreation       `json:"cache_creation"`
	OutputTokens             int                 `json:"output_tokens"`
	OutputTokensDetails      OutputTokensDetails `json:"output_tokens_details"`
	ServiceTier              string              `json:"service_tier"`
	InferenceGeo             string              `json:"inference_geo"`
	Speed                    string              `json:"speed"`
}

func (svc Service) CreateMessage(ctx context.Context, req CreateMessageRequest) (*CreateMessageResponse, error) {
	var res CreateMessageResponse
	err := requests.
		URL("https://api.anthropic.com/v1/messages").
		Client(svc.Client).
		Header("anthropic-version", "2026-10-21").
		Header("x-api-key", svc.Key).
		BodyJSON(req).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
