// Package google has Google API client stuff
package google

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"io"
)

func AddFlags(fl *flag.FlagSet) func() *Service {
	var gsvc Service
	// Using a crazy Base64+GZIP because storing JSON containing \n in
	//an env var breaks a lot for some reason
	fl.Func("google-json", "GZIP Base64 JSON `credentials` for Google",
		func(s string) error {
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
		})
	fl.StringVar(&gsvc.viewID, "ga-view-id", "", "view `ID` for Google Analytics")
	fl.StringVar(&gsvc.driveID, "google-drive-id", "", "`ID` for shared Google Drive")
	return func() *Service {
		return &gsvc
	}
}

type Service struct {
	cert    []byte
	viewID  string
	driveID string
}

func (gsvc *Service) ConfigureCert(s string) error {
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
