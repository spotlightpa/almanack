package herokuapi_test

import (
	"bytes"
	"flag"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/herokuapi"
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
	if err != nil {
		t.Fatal("err", err)
	}
	var buf bytes.Buffer
	l := log.New(&buf, "", log.LstdFlags)
	err = conf.Configure(l, map[string]string{
		"db": "TEST_KEY",
		"xx": "MISSING_VAL",
	})
	// t.Log(buf.String())
	if err != nil {
		t.Fatal("err", err)
	}
	if *dbstr == "" {
		t.Fatal("no TEST_KEY")
	}
	if *xxstr != "initial" {
		t.Fatalf("overwrote missing value")
	}
}
