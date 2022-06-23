package herokuapi_test

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/herokuapi"
	"github.com/spotlightpa/almanack/pkg/common"
)

func TestHerokuAPI(t *testing.T) {
	apiKey := os.Getenv("HEROKU_API_KEY")
	appName := os.Getenv("HEROKU_APP_NAME")
	http.DefaultClient.Transport = requests.Replay("testdata")
	if os.Getenv("HEROKU_RECORD_REQUEST") != "" {
		http.DefaultClient.Transport = requests.Record(nil, "testdata")
	}
	t.Cleanup(func() {
		http.DefaultClient.Transport = nil
	})
	if apiKey == "" {
		apiKey = "testing123"
	}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	conf := herokuapi.AddFlags(fs)
	dbstr := fs.String("db", "", "")
	xxstr := fs.String("xx", "initial", "")
	err := fs.Parse([]string{
		"-heroku-api-key", apiKey,
		"-heroku-app-name", appName})
	be.NilErr(t, err)
	var buf bytes.Buffer
	common.Logger = log.New(&buf, "", log.LstdFlags)
	err = conf.Configure(map[string]string{
		"db": "TEST_KEY",
		"xx": "MISSING_VAL",
	})
	be.NilErr(t, err)
	be.Nonzero(t, *dbstr)
	be.Equal(t, "initial", *xxstr)
}
