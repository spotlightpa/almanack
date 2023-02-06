package gdocs

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/nkotb/pkg/testfile"
	"github.com/spotlightpa/nkotb/pkg/xhtml"
	"google.golang.org/api/docs/v1"
)

func TestConvert(t *testing.T) {
	testfile.GlobRun(t, "testdata/*.html", func(path string, t *testing.T) {
		bareName := strings.TrimSuffix(path, ".html")
		want := testfile.Read(t, path)

		s := testfile.Read(t, bareName+".json")

		var doc docs.Document
		be.NilErr(t, json.Unmarshal([]byte(s), &doc))
		n := Convert(&doc)
		got := xhtml.ToString(n)

		be.Debug(t, func() {
			badname := bareName + "-bad.html"
			testfile.Write(t, badname, got)
		})
		be.Equal(t, string(want), got)
	})
}
