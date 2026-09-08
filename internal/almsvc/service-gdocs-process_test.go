package almsvc

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/utils/must"
	"golang.org/x/net/html"
)

func TestProcessDocHTML(t *testing.T) {
	testfile.Run(t, "testdata/processDocHTML/*/doc.html", func(be assert.TB, path string) {
		input := testfile.Read(be, path)
		doc := must.Get(html.Parse(strings.NewReader(input)))
		metadata, embeds, intDoc, richText, rawHTML, md, warnings := processDocHTML(doc)

		dir := filepath.Dir(path)

		intermediateDoc := xhtml.OuterHTML(intDoc)
		richTextStr := xhtml.OuterHTML(richText)
		rawHTMLStr := xhtml.OuterHTML(rawHTML)

		testfile.Equalish(be, filepath.Join(dir, "intermediate.html"), intermediateDoc)
		testfile.Equalish(be, filepath.Join(dir, "rich.html"), richTextStr)
		testfile.Equalish(be, filepath.Join(dir, "raw.html"), rawHTMLStr)
		testfile.Equalish(be, filepath.Join(dir, "article.md"), md)
		testfile.EqualJSON(be, filepath.Join(dir, "metadata.json"), metadata)
		testfile.EqualJSON(be, filepath.Join(dir, "embeds.json"), embeds)
		testfile.EqualJSON(be, filepath.Join(dir, "warnings.json"), warnings)
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
