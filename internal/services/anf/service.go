package anf

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/carlmjohnson/requests"
)

type Service struct {
	ChannelID, Key, Secret string
	*http.Client
}

func AddFlags(fl *flag.FlagSet) (svc *Service) {
	svc = new(Service)
	fl.StringVar(&svc.ChannelID, "apple-news-channel-id", "", "`channel id` for Apple News Publisher")
	fl.StringVar(&svc.Key, "apple-news-key", "", "`key` for Apple News Publisher")
	fl.StringVar(&svc.Secret, "apple-news-secret", "", "`secret` for Apple News Publisher")
	svc.Client = &http.Client{
		Transport: requests.ErrorTransport(errors.New("apple news client not configured")),
	}
	return svc
}

type ServiceErrorResponse struct {
	Errors []struct {
		Code  string `json:"code"`
		Value string `json:"value"`
	} `json:"errors"`
}

func (svc *Service) client() *http.Client {
	cl2 := *svc.Client
	cl2.Transport = HHMacTransport(svc.Key, svc.Secret, cl2.Transport)
	return &cl2
}

func (svc *Service) Create(ctx context.Context, a *Article, sections []string) (*Response, error) {
	type ArticleLinksRequest struct {
		Sections []string `json:"sections"`
	}
	type Data struct {
		Links ArticleLinksRequest `json:"links"`
	}
	metadata := struct {
		Data Data `json:"data"`
	}{Data{
		Links: ArticleLinksRequest{Sections: sections},
	}}
	metadataB, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	var res Response
	var errDetails ServiceErrorResponse
	err = requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/articles", svc.ChannelID).
		Client(svc.client()).
		Config(requests.BodyMultipart("", func(multi *multipart.Writer) error {
			if err := writeFile(multi, "metadata.json", "application/json", metadataB); err != nil {
				return err
			}
			if err := writeFile(multi, "article.json", "application/json", data); err != nil {
				return err
			}
			return nil
		})).
		ErrorJSON(&errDetails).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v; %w", errDetails, err)
	}
	return &res, nil
}

func (svc *Service) Update(ctx context.Context, a *Article, appleID, revision string, sections []string) (*Response, error) {
	type ArticleLinksRequest struct {
		Sections []string `json:"sections"`
	}
	type Data struct {
		Revision string              `json:"revision"`
		Links    ArticleLinksRequest `json:"links"`
	}
	metadata := struct {
		Data Data `json:"data"`
	}{Data{
		Revision: revision,
		Links:    ArticleLinksRequest{Sections: sections},
	}}
	metadataB, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	var res Response
	var errDetails ServiceErrorResponse
	err = requests.
		URL("https://news-api.apple.com").
		Pathf("/articles/%s", appleID).
		Client(svc.client()).
		Config(requests.BodyMultipart("", func(multi *multipart.Writer) error {
			if err := writeFile(multi, "metadata", "application/json", metadataB); err != nil {
				return err
			}
			if err := writeFile(multi, "article.json", "application/json", data); err != nil {
				return err
			}
			return nil
		})).
		ErrorJSON(&errDetails).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v; %w", errDetails, err)
	}
	return &res, nil
}

func (svc *Service) ReadChannel(ctx context.Context) (any, error) {
	var data any
	var errDetails ServiceErrorResponse
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/", svc.ChannelID).
		Client(svc.client()).
		ErrorJSON(&errDetails).
		ToJSON(&data).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v; %w", errDetails, err)
	}
	return data, nil
}

func (svc *Service) List(ctx context.Context) (any, error) {
	var data any
	var errDetails ServiceErrorResponse
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/articles", svc.ChannelID).
		Client(svc.client()).
		ErrorJSON(&errDetails).
		ToJSON(&data).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v; %w", errDetails, err)
	}
	return data, nil
}

func (svc *Service) ListSections(ctx context.Context) (*ListSectionResponse, error) {
	var data ListSectionResponse
	var errDetails ServiceErrorResponse
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/sections", svc.ChannelID).
		Client(svc.client()).
		ErrorJSON(&errDetails).
		ToJSON(&data).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v; %w", errDetails, err)
	}
	return &data, nil
}

func (svc *Service) ReadArticle(ctx context.Context, articleID string) (*Response, error) {
	var res Response
	var errDetails ServiceErrorResponse
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/articles/%s", articleID).
		Client(svc.client()).
		ErrorJSON(&errDetails).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v; %w", errDetails, err)
	}
	return &res, nil
}

func writeFile(multi *multipart.Writer, name, contentType string, content []byte) error {
	h := make(textproto.MIMEHeader)
	disposition := fmt.Sprintf(`form-data; filename=%s; size=%d`, name, len(content))
	h.Set("Content-Disposition", disposition)
	h.Set("Content-Type", contentType)
	w, err := multi.CreatePart(h)
	if err != nil {
		return err
	}
	_, err = w.Write(content)
	if err != nil {
		return err
	}
	return nil
}
