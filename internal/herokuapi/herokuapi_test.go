package herokuapi_test

import (
	"os"
	"testing"

	"github.com/spotlightpa/almanack/internal/herokuapi"
)

func TestHerokuAPI(t *testing.T) {
	apiKey := os.Getenv("HEROKU_API_KEY")
	addOn := os.Getenv("HEROKU_ADD_ON_ID")
	if apiKey == "" {
		t.Skip("no API key specified")
	}
	if addOn == "" {
		t.Skip("no add on specified")
	}
	connURL, err := herokuapi.Request(apiKey, addOn)
	if err != nil {
		t.Errorf("expected err == nil: err == %v", err)
	}
	if connURL == "" {
		t.Errorf(`expected connURL != ""`)
	}
}
