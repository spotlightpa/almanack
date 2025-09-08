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
}

func AddFlags(fl *flag.FlagSet) (svc *Service) {
	svc = new(Service)
	fl.StringVar(&svc.ChannelID, "apple-news-channel-id", "", "`channel id` for Apple News Publisher")
	fl.StringVar(&svc.Key, "apple-news-key", "", "`key` for Apple News Publisher")
	fl.StringVar(&svc.Secret, "apple-news-secret", "", "`secret` for Apple News Publisher")
	return svc
}

func (svc *Service) Publish(ctx context.Context, cl *http.Client, a *Article) error {
	cl2 := *cl
	cl2.Transport = HHMacTransport(svc.Key, svc.Secret, cl.Transport)
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/articles", svc.ChannelID).
		Client(&cl2).
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
		Fetch(ctx)
	return err
}

func (svc *Service) Read(ctx context.Context, cl *http.Client) (any, error) {
	cl2 := *cl
	cl2.Transport = HHMacTransport(svc.Key, svc.Secret, cl.Transport)
	var data any
	var errDetails struct {
		Errors []struct{ Code string }
	}
	err := requests.
		URL("https://news-api.apple.com").
		Pathf("/channels/%s/", svc.ChannelID).
		Client(&cl2).
		ErrorJSON(&errDetails).
		ToJSON(&data).
		Fetch(ctx)
	if err != nil {
		return nil, fmt.Errorf("service Apple News error: %v", errDetails)
	}
	return data, nil
}
