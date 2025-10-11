package google

import (
	"cmp"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestTranslate(t *testing.T) {
	almlog.UseTestLogger(t)
	svc := Service{}
	svc.projectID = cmp.Or(os.Getenv("ALMANACK_GOOGLE_PROJECT_ID"), "1")
	ctx := t.Context()
	cl := *http.DefaultClient
	cl.Transport = reqtest.Replay("testdata")
	if os.Getenv("ALMANACK_GOOGLE_TEST_RECORD_REQUEST") != "" {
		gcl, err := svc.TranslateClient(ctx)
		be.NilErr(t, err)
		cl.Transport = reqtest.Record(gcl.Transport, "testdata")
	}
	translated, err := svc.Translate(ctx, &cl, "Hello, World!", "text/plain")
	be.NilErr(t, err)
	be.Equal(t, "¡Hola Mundo!", translated)
}
