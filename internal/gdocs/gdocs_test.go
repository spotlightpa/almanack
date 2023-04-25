package gdocs

import (
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
		var doc docs.Document
		testfile.ReadJSON(t, path, &doc)

		n := Convert(&doc)
		got := xhtml.ToString(n)

		bareName := strings.TrimSuffix(path, ".json")
		testfile.Equal(t, bareName+".html", got)
	})
}

func TestFullConvert(t *testing.T) {
	t.Parallel()
	testfile.GlobRun(t, "testdata/*.json", func(path string, t *testing.T) {
		var doc docs.Document
		testfile.ReadJSON(t, path, &doc)

		n := Convert(&doc)
		got, err := blocko.MinifyAndBlockize(xhtml.ToString(n))
		be.NilErr(t, err)

		bareName := strings.TrimSuffix(path, ".json")
		testfile.Equal(t, bareName+".md", got)
	})
}

func BenchmarkConvert(b *testing.B) {
	want := testfile.Read(be.Relaxed(b), "testdata/privacy.html")
	var got *html.Node

	var doc docs.Document
	testfile.ReadJSON(b, "testdata/privacy.json", &doc)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		got = Convert(&doc)
	}
	be.Equal(b, want, xhtml.ToString(got))
}

func BenchmarkFullConvert(b *testing.B) {
	want := testfile.Read(be.Relaxed(b), "testdata/privacy.md")
	var got string

	var doc docs.Document
	testfile.ReadJSON(b, "testdata/privacy.json", &doc)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := Convert(&doc)
		var err error
		got, err = blocko.MinifyAndBlockize(xhtml.ToString(n))
		be.NilErr(b, err)
	}
	be.Equal(b, want, got)
}
