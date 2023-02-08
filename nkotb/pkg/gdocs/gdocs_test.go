package gdocs

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/nkotb/pkg/blocko"
	"github.com/spotlightpa/nkotb/pkg/testfile"
	"github.com/spotlightpa/nkotb/pkg/xhtml"
	"google.golang.org/api/docs/v1"
)

func TestConvert(t *testing.T) {
	testfile.GlobRun(t, "testdata/*.json", func(path string, t *testing.T) {
		bareName := strings.TrimSuffix(path, ".json")
		want := testfile.Read(be.Relaxed(t), bareName+".html")

		s := testfile.Read(t, path)

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

func TestFullConvert(t *testing.T) {
	testfile.GlobRun(t, "testdata/*.json", func(path string, t *testing.T) {
		bareName := strings.TrimSuffix(path, ".json")
		want := testfile.Read(be.Relaxed(t), bareName+".md")

		s := testfile.Read(t, path)

		var doc docs.Document
		be.NilErr(t, json.Unmarshal([]byte(s), &doc))
		n := Convert(&doc)
		got, err := blocko.HTMLToMarkdown(xhtml.ToString(n))
		be.NilErr(t, err)

		be.Debug(t, func() {
			badname := bareName + "-bad.md"
			testfile.Write(t, badname, got)
		})
		be.Equal(t, string(want), got)
	})
}
