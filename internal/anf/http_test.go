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
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/anf"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestHMACSignRequest(t *testing.T) {
	testfile.Run(t, "testdata/req.*.raw", func(t *testing.T, match string) {
		synctest.Test(t, func(t *testing.T) {
			in := testfile.Read(t, match)
			buf := bufio.NewReader(strings.NewReader(in))
			req, err := http.ReadRequest(buf)
			be.NilErr(t, err)

			now := time.Now()
			be.NilErr(t, anf.HHMACSignRequest(req, "key", "aGVsbG8=", now))
			signed, err := httputil.DumpRequest(req, true)
			be.NilErr(t, err)
			testfile.Equalish(t, testfile.Ext(match, "signed"), string(signed))
		})
	})
}

func TestRead(t *testing.T) {
	almlog.UseTestLogger(t)
	svc := anf.Service{
		ChannelID: "abc",
		Key:       "123",
		Secret:    "aGVsbG8=",
		Client: &http.Client{
			Transport: reqtest.Replay("testdata/api/"),
		},
	}
	synctest.Test(t, func(t *testing.T) {
		data, err := svc.ReadChannel(t.Context())
		be.NilErr(t, err)
		be.Nonzero(t, data)
	})
}
