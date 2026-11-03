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
	testfile.Run(t, "testdata/req.*.raw", func(be assert.TB, match string) {
		synctest.Test(be.TB.(*testing.T), func(t *testing.T) {
			be := assert.FailsNow(t)
			in := testfile.Read(t, match)
			buf := bufio.NewReader(strings.NewReader(in))
			req := be.OK(http.ReadRequest(buf))

			now := time.Now()
			be.NilError(anf.HHMACSignRequest(req, "key", "aGVsbG8=", now))
			signed := be.OK(httputil.DumpRequest(req, true))
			testfile.Equalish(t, testfile.Ext(match, "signed"), string(signed))
		})
	})
}

func TestService(t *testing.T) {
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
		be := assert.FailsNow(t)
		data := be.OK(svc.ReadChannel(t.Context()))
		be.Truthy(data)
		sections := be.OK(svc.ListSections(t.Context()))
		be.Truthy(sections)
		// Should have at least default channel
		be.Truthy(sections.ToMap()[""])
	})
}
