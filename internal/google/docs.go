package google

import (
	"context"
	"net/http"

	"github.com/spotlightpa/almanack/pkg/almlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (gsvc *Service) GDocsClient(ctx context.Context) (cl *http.Client, err error) {
	if len(gsvc.cert) == 0 {
		l := almlog.FromContext(ctx)
		l.Warn("using default Google credentials")
		cl, err = google.DefaultClient(ctx,
			"https://www.googleapis.com/auth/documents.readonly",
			"https://www.googleapis.com/auth/drive.readonly",
		)
		return
	}
	creds, err := google.CredentialsFromJSON(
		ctx, gsvc.cert,
		"https://www.googleapis.com/auth/documents.readonly",
		"https://www.googleapis.com/auth/drive.readonly",
	)
	if err != nil {
		return
	}
	cl = oauth2.NewClient(ctx, creds.TokenSource)
	return
}
