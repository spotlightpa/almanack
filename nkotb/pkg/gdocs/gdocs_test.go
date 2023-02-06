package gdocs

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/nkotb/pkg/xhtml"
	"google.golang.org/api/docs/v1"
)

func TestConvert(t *testing.T) {
	for _, tc := range []string{
		"example",
	} {
		want, err := os.ReadFile(fmt.Sprintf("testdata/%s.html", tc))
		be.NilErr(t, err)

		b, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", tc))
		be.NilErr(t, err)

		var doc docs.Document
		be.NilErr(t, json.Unmarshal(b, &doc))
		n := Convert(&doc)
		got := xhtml.ToString(n)

		be.Debug(t, func() {
			badname := fmt.Sprintf("testdata/%s-bad.html", tc)
			os.WriteFile(badname, []byte(got), 0644)
		})
		be.Equal(t, string(want), got)
	}

}
