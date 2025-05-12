package almanack

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/must"
	"golang.org/x/net/html"
)

func TestProcessDocHTML(t *testing.T) {
	testfile.Run(t, "testdata/processDocHTML/*/doc.html", func(t *testing.T, path string) {
		input := testfile.Read(t, path)
		doc := must.Get(html.Parse(strings.NewReader(input)))
		metadata, embeds, richText, rawHTML, md, warnings := processDocHTML(doc)

		dir := filepath.Dir(path)

		richTextStr := xhtml.OuterHTML(richText)
		rawHTMLStr := xhtml.OuterHTML(rawHTML)

		rt := be.Relaxed(t)
		testfile.Equalish(rt, filepath.Join(dir, "rich.html"), richTextStr)
		testfile.Equalish(rt, filepath.Join(dir, "raw.html"), rawHTMLStr)
		testfile.Equalish(rt, filepath.Join(dir, "article.md"), md)
		testfile.EqualJSON(rt, filepath.Join(dir, "metadata.json"), metadata)
		testfile.EqualJSON(rt, filepath.Join(dir, "embeds.json"), embeds)
		testfile.EqualJSON(rt, filepath.Join(dir, "warnings.json"), warnings)
	})
}

func BenchmarkProcessDocHTML(b *testing.B) {
	input := testfile.Read(b, "testdata/processDocHTML/SPLEX23ERR/doc.html")
	doc := must.Get(html.Parse(strings.NewReader(input)))
	b.ResetTimer()
	for range b.N {
		processDocHTML(xhtml.Clone(doc))
	}
}
