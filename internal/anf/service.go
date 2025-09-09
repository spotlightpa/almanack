package anf

import (
	"context"
	"encoding/json"
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
	return svc
}

func (svc *Service) Publish(ctx context.Context, a *Article) (*Response, error) {
	var errDetails struct {
		Errors []struct{ Code string }
	}
	var res Response
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/articles", svc.ChannelID).
		Client(svc.client()).
		ErrorJSON(&errDetails).
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
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v", errDetails)
	}
	return &res, nil
}

func (svc *Service) Update(ctx context.Context, a *Article, appleID, revision string) (*Response, error) {
	cl2 := *svc.Client
	cl2.Transport = HHMacTransport(svc.Key, svc.Secret, cl2.Transport)
	var errDetails struct {
		Errors []struct{ Code string }
	}
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
	err = requests.
		URL("https://news-api.apple.com").
		Pathf("/articles/%s", appleID).
		Client(svc.client()).
		ErrorJSON(&errDetails).
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
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v", errDetails)
	}
	return &res, nil
}

func (svc *Service) Read(ctx context.Context) (any, error) {
	var data any
	var errDetails struct {
		Errors []struct{ Code string }
	}
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/", svc.ChannelID).
		Client(svc.client()).
		ErrorJSON(&errDetails).
		ToJSON(&data).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v", errDetails)
	}
	return data, nil
}

func (svc *Service) List(ctx context.Context) (any, error) {
	var data any
	var errDetails struct {
		Errors []struct{ Code string }
	}
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/articles", svc.ChannelID).
		Client(svc.client()).
		ErrorJSON(&errDetails).
		ToJSON(&data).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v", errDetails)
	}
	return data, nil
}

func (svc *Service) client() *http.Client {
	cl2 := *svc.Client
	cl2.Transport = HHMacTransport(svc.Key, svc.Secret, cl2.Transport)
	return &cl2
}
