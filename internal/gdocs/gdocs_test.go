package gdocs

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/testfile"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
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
		be.Equal(t, want, got)
	})
}

func TestFullConvert(t *testing.T) {
	t.Parallel()
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
		be.Equal(t, want, got)
	})
}

func BenchmarkConvert(b *testing.B) {
	want := testfile.Read(be.Relaxed(b), "testdata/privacy.html")
	s := testfile.Read(b, "testdata/privacy.json")
	var got *html.Node
	var doc docs.Document
	be.NilErr(b, json.Unmarshal([]byte(s), &doc))

	for i := 0; i < b.N; i++ {
		got = Convert(&doc)
	}
	be.Equal(b, want, xhtml.ToString(got))
}

func BenchmarkFullConvert(b *testing.B) {
	want := testfile.Read(be.Relaxed(b), "testdata/privacy.md")
	s := testfile.Read(b, "testdata/privacy.json")
	var got string

	var doc docs.Document
	be.NilErr(b, json.Unmarshal([]byte(s), &doc))

	for i := 0; i < b.N; i++ {
		n := Convert(&doc)
		var err error
		got, err = blocko.HTMLToMarkdown(xhtml.ToString(n))
		be.NilErr(b, err)
	}
	be.Equal(b, want, got)
}
