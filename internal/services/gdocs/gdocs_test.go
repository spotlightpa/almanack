package gdocs

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/convert/blocko"
	"golang.org/x/net/html"
	"google.golang.org/api/docs/v1"
)

func TestConvert(t *testing.T) {
	testfile.Run(t, "testdata/*.json", func(t *testing.T, path string) {
		var doc docs.Document
		testfile.ReadJSON(t, path, &doc)

		n := Convert(&doc)
		got := xhtml.OuterHTML(n)

		testfile.Equalish(t, testfile.Ext(path, ".html"), got)
	})
}

func TestFullConvert(t *testing.T) {
	t.Parallel()
	testfile.Run(t, "testdata/*.json", func(t *testing.T, path string) {
		var doc docs.Document
		testfile.ReadJSON(t, path, &doc)

		n := Convert(&doc)
		got := assert.FailNow(t).OK(blocko.MinifyAndBlockize(xhtml.OuterHTML(n)))
		testfile.Equalish(t, testfile.Ext(path, ".md"), got)
	})
}

func BenchmarkConvert(b *testing.B) {
	want := testfile.Read(b, "testdata/privacy.html")
	var got *html.Node

	var doc docs.Document
	testfile.ReadJSON(b, "testdata/privacy.json", &doc)
	b.ResetTimer()

	for b.Loop() {
		got = Convert(&doc)
	}
	assert.FailNow(b).Equal(xhtml.OuterHTML(got), want)
}

func BenchmarkFullConvert(b *testing.B) {
	be := assert.Continue(b)
	want := testfile.Read(b, "testdata/privacy.md")
	var got string

	var doc docs.Document
	testfile.ReadJSON(b, "testdata/privacy.json", &doc)
	b.ResetTimer()

	for b.Loop() {
		n := Convert(&doc)
		got = be.OK(blocko.MinifyAndBlockize(xhtml.OuterHTML(n)))
	}
	be.Equal(got, want)
}
