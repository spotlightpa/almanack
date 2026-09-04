package mailchimp_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/mailchimp"
)

func TestSendEmail(t *testing.T) {
	almlog.UseTestLogger(t)

	cl := *http.DefaultClient
	cl.Transport = reqtest.Replay("testdata/sendemail")
	apiKey := os.Getenv("ALMANACK_MC_TEST_API_KEY")
	listID := os.Getenv("ALMANACK_MC_TEST_LISTID")

	if os.Getenv("RECORD") != "" {
		cl.Transport = reqtest.Caching(nil, "testdata/sendemail")
	}
	v3 := mailchimp.NewV3(apiKey, listID, &cl)
	assert.FailNow(t).Zero(v3.SendEmail(t.Context(), "Test message", "Hello, World!"))
}
