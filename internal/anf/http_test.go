package anf_test

import (
	"bufio"
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/spotlightpa/almanack/internal/anf"
)

func TestHMACSignRequest(t *testing.T) {
	testfile.Run(t, "testdata/req.*.raw", func(t *testing.T, match string) {
		synctest.Test(t, func(t *testing.T) {
			in := testfile.Read(t, match)
			buf := bufio.NewReader(strings.NewReader(in))
			req, err := http.ReadRequest(buf)
			be.NilErr(t, err)

			now := time.Now()
			be.NilErr(t, anf.HHMACSignRequest(req, "key", "abc123", now))
			signed, err := httputil.DumpRequest(req, true)
			be.NilErr(t, err)
			testfile.Equalish(t, testfile.Ext(match, "signed"), string(signed))
		})
	})
}
