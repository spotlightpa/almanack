package mailchimp_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestSendEmail(t *testing.T) {
	almlog.UseTestLogger(t)

	cl := *http.DefaultClient
	cl.Transport = requests.Replay("testdata/sendemail")
	apiKey := os.Getenv("ALMANACK_MC_TEST_API_KEY")
	listID := os.Getenv("ALMANACK_MC_TEST_LISTID")

	if os.Getenv("RECORD") != "" {
		cl.Transport = requests.Caching(nil, "testdata/sendemail")
	}
	v3 := mailchimp.NewV3(apiKey, listID, &cl)
	err := v3.SendEmail(context.Background(), "Test message", "Hello, World!")
	be.NilErr(t, err)
}
