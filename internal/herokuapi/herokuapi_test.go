package herokuapi_test

import (
	"bytes"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/spotlightpa/almanack/internal/herokuapi"
)

func TestHerokuAPI(t *testing.T) {
	apiKey := os.Getenv("HEROKU_API_KEY")
	appName := os.Getenv("HEROKU_APP_NAME")
	if apiKey == "" {
		t.Skip("no API key specified")
	}
	if appName == "" {
		t.Skip("no app name specified")
	}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	conf := herokuapi.ConfigureFlagSet(fs)
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
		"db": "DATABASE_URL",
		"xx": "MISSING_VAL",
	})
	// t.Log(buf.String())
	if err != nil {
		t.Fatal("err", err)
	}
	if *dbstr == "" {
		t.Fatal("no DATABASE_URL")
	}
	if *xxstr != "initial" {
		t.Fatalf("overwrote missing value")
	}
}
