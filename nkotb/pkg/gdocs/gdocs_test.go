package gdocs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"golang.org/x/net/html"
	"google.golang.org/api/docs/v1"
)

func nilErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("want nil; got %v", err)
	}
}

func TestConvert(t *testing.T) {
	for _, tc := range []string{
		"example",
	} {
		b, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", tc))
		nilErr(t, err)

		var doc docs.Document
		nilErr(t, json.Unmarshal(b, &doc))
		n := Convert(&doc)
		var buf bytes.Buffer
		nilErr(t, html.Render(&buf, n))

		want, err := os.ReadFile(fmt.Sprintf("testdata/%s.html", tc))
		nilErr(t, err)

		if buf.String() != string(want) {
			badname := fmt.Sprintf("testdata/%s-bad.html", tc)
			os.WriteFile(badname, buf.Bytes(), 0644)
			t.Fatalf("unexpected output see %s", badname)
		}
	}

}
