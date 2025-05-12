package gdocs

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/blocko"
	"golang.org/x/net/html"
	"google.golang.org/api/docs/v1"
)

func TestConvert(t *testing.T) {
	testfile.Run(t, "testdata/*.json", func(t *testing.T, path string) {
		var doc docs.Document
		testfile.ReadJSON(t, path, &doc)

		n := Convert(&doc)
		got := xhtml.OuterHTML(n)

		testfile.Equalish(be.Relaxed(t), testfile.Ext(path, ".html"), got)
	})
}

func TestFullConvert(t *testing.T) {
	t.Parallel()
	testfile.Run(t, "testdata/*.json", func(t *testing.T, path string) {
		var doc docs.Document
		testfile.ReadJSON(t, path, &doc)

		n := Convert(&doc)
		got, err := blocko.MinifyAndBlockize(xhtml.OuterHTML(n))
		be.NilErr(t, err)

		testfile.Equalish(t, testfile.Ext(path, ".md"), got)
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
	be.Equal(b, want, xhtml.OuterHTML(got))
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
		got, err = blocko.MinifyAndBlockize(xhtml.OuterHTML(n))
		be.NilErr(b, err)
	}
	be.Equal(b, want, got)
}
