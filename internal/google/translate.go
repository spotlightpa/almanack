package google

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/requests"
)

func (gsvc *Service) TranslateClient(ctx context.Context) (cl *http.Client, err error) {
	return gsvc.client(ctx, `https://www.googleapis.com/auth/cloud-translation`)
}

func (gsvc *Service) Translate(ctx context.Context, cl *http.Client, source, contentType string) (string, error) {
	type Req struct {
		Contents           []string `json:"contents"`
		SourceLanguageCode string   `json:"sourceLanguageCode"`
		TargetLanguageCode string   `json:"targetLanguageCode"`
		MimeType           string   `json:"mimeType"`
	}
	type Translations struct {
		TranslatedText string `json:"translatedText"`
	}
	type Res struct {
		Translations []Translations `json:"translations"`
	}
	req := Req{
		Contents:           []string{source},
		SourceLanguageCode: "en-US",
		TargetLanguageCode: "es",
		MimeType:           contentType,
	}
	var res Res
	if err := requests.
		URL("https://translate.googleapis.com").
		Pathf("/v3/projects/%s:translateText", gsvc.projectID).
		Client(cl).
		BodyJSON(&req).
		ToJSON(&res).
		Fetch(ctx); err != nil {
		return "", fmt.Errorf("could not translate: %w", err)
	}
	if len(res.Translations) != 1 {
		return "", fmt.Errorf("unexpected response array: %v", res.Translations)
	}
	return res.Translations[0].TranslatedText, nil
}
