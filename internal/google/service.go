package google

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"io"

	"github.com/spotlightpa/almanack/pkg/common"
)

func FlagVar(fl *flag.FlagSet) func(l common.Logger) *Service {
	var ga Service
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
			ga.cert, err = io.ReadAll(g)
			if err != nil {
				return err
			}
			return nil
		})
	fl.StringVar(&ga.viewID, "ga-view-id", "", "view `ID` for Google Analytics")
	return func(l common.Logger) *Service {
		ga.l = l
		return &ga
	}
}

type Service struct {
	cert   []byte
	l      common.Logger
	viewID string
}
