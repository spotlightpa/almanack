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

func (svc *Service) Create(ctx context.Context, a *Article) (*Response, error) {
	var res Response
	var errDetails ServiceErrorResponse
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/articles", svc.ChannelID).
		Client(svc.client()).
		Config(requests.BodyMultipart("", func(multi *multipart.Writer) error {
			data, err := json.Marshal(a)
			if err != nil {
				return err
			}
			h := make(textproto.MIMEHeader)
			disposition := fmt.Sprintf(`form-data; filename=article.json; size=%d`, len(data))
			h.Set("Content-Disposition", disposition)
			h.Set("Content-Type", "application/json")
			w, err := multi.CreatePart(h)
			if err != nil {
				return err
			}
			_, err = w.Write(data)
			return err
		})).
		ErrorJSON(&errDetails).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v; %w", errDetails, err)
	}
	return &res, nil
}

func (svc *Service) Update(ctx context.Context, a *Article, appleID, revision string) (*Response, error) {
	cl2 := *svc.Client
	cl2.Transport = HHMacTransport(svc.Key, svc.Secret, cl2.Transport)
	type Data struct {
		Revision string `json:"revision"`
	}
	metadata := struct {
		Data Data `json:"data"`
	}{Data{Revision: revision}}
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
			{
				h := make(textproto.MIMEHeader)
				disposition := fmt.Sprintf(`form-data; name=metadata; size=%d`, len(metadataB))
				h.Set("Content-Disposition", disposition)
				h.Set("Content-Type", "application/json")
				w, err := multi.CreatePart(h)
				if err != nil {
					return err
				}
				if _, err = w.Write(metadataB); err != nil {
					return err
				}
			}
			{
				h := make(textproto.MIMEHeader)
				disposition := fmt.Sprintf(`form-data; filename=article.json; size=%d`, len(data))
				h.Set("Content-Disposition", disposition)
				h.Set("Content-Type", "application/json")
				w, err := multi.CreatePart(h)
				if err != nil {
					return err
				}
				if _, err = w.Write(data); err != nil {
					return err
				}
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
