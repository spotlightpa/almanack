package anf_test

import (
	"bufio"
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/anf"
)

func TestHMACSignRequest(t *testing.T) {
	testfile.Run(t, "testdata/req.*.raw", func(t *testing.T, match string) {
		be := assert.FailNow(t)
		synctest.Test(t, func(t *testing.T) {
			in := testfile.Read(t, match)
			buf := bufio.NewReader(strings.NewReader(in))
			req, err := http.ReadRequest(buf)
			be.Zero(err)

			now := time.Now()
			be.Zero(anf.HHMACSignRequest(req, "key", "aGVsbG8=", now))
			signed, err := httputil.DumpRequest(req, true)
			be.Zero(err)
			testfile.Equalish(t, testfile.Ext(match, "signed"), string(signed))
		})
	})
}

func TestService(t *testing.T) {
	be := assert.FailNow(t)
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
		be.Zero(err)
		be.NotZero(data)
		sections, err := svc.ListSections(t.Context())
		be.Zero(err)
		be.NotZero(sections)
		// Should have at least default channel
		be.NotZero(sections.ToMap()[""])
	})
}
