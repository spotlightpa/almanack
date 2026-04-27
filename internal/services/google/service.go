// Package google has Google API client stuff
package google

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"flag"
	"io"
	"net/http"
	"sync"

	"github.com/spotlightpa/almanack/pkg/almlog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Service struct {
	certMU     sync.RWMutex
	cert       []byte
	mockClient *http.Client

	viewID    string
	driveID   string
	projectID string
}

func (gsvc *Service) AddFlags(fl *flag.FlagSet) {
	// Using a crazy Base64+GZIP because storing JSON containing \n in
	//an env var breaks a lot for some reason
	fl.Func("google-json", "GZIP Base64 JSON `credentials` for Google", gsvc.ConfigureCert)
	fl.StringVar(&gsvc.viewID, "ga-view-id", "", "view `ID` for Google Analytics")
	fl.StringVar(&gsvc.driveID, "google-drive-id", "", "`ID` for shared Google Drive")
	fl.StringVar(&gsvc.projectID, "google-project-id", "", "`ID` for Google Cloud project")
}

func (gsvc *Service) HasCert() bool {
	gsvc.certMU.RLock()
	defer gsvc.certMU.RUnlock()
	return len(gsvc.cert) > 0
}

func (gsvc *Service) SetMockClient(cl *http.Client) {
	gsvc.certMU.Lock()
	defer gsvc.certMU.Unlock()
	gsvc.mockClient = cl
}

func (gsvc *Service) ConfigureCert(s string) error {
	gsvc.certMU.Lock()
	defer gsvc.certMU.Unlock()

	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	g, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer g.Close()

	gsvc.cert, err = io.ReadAll(g)
	if err != nil {
		return err
	}
	return nil
}

func (gsvc *Service) client(ctx context.Context, scopes ...string) (cl *http.Client, err error) {
	gsvc.certMU.RLock()
	defer gsvc.certMU.RUnlock()

	if gsvc.mockClient != nil {
		return gsvc.mockClient, nil
	}

	if len(gsvc.cert) == 0 {
		l := almlog.FromContext(ctx)
		l.WarnContext(ctx, "using default Google credentials")
		cl, err = google.DefaultClient(ctx, scopes...)
		return
	}
	creds, err := google.CredentialsFromJSON(ctx, gsvc.cert, scopes...)
	if err != nil {
		return
	}
	cl = oauth2.NewClient(ctx, creds.TokenSource)
	return
}
