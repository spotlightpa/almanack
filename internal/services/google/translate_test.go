package google

import (
	"cmp"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"net/http"
	"os"
	"testing"
)

func TestTranslate(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	svc := Service{
		projectID: cmp.Or(os.Getenv("ALMANACK_GOOGLE_PROJECT_ID"), "1")}
	ctx := t.Context()
	cl := *http.DefaultClient
	cl.Transport = reqtest.Replay("testdata")
	if os.Getenv("ALMANACK_GOOGLE_TEST_RECORD_REQUEST") != "" {
		gcl, err := svc.TranslateClient(ctx)
		be.Zero(err)
		cl.Transport = reqtest.Record(gcl.Transport, "testdata")
	}
	translated, err := svc.Translate(ctx, &cl, "text/plain", "Hello, World!")
	be.Zero(err)
	be.EqualLength(translated, 1)
	be.Equal(translated[0], "¡Hola Mundo!")
}
