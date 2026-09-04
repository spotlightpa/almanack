package almsvc

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/earthboundkid/assert/testfile"
	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/utils/must"
	"golang.org/x/net/html"
)

func TestProcessDocHTML(t *testing.T) {
	testfile.Run(t, "testdata/processDocHTML/*/doc.html", func(t *testing.T, path string) {
		input := testfile.Read(t, path)
		doc := must.Get(html.Parse(strings.NewReader(input)))
		metadata, embeds, intDoc, richText, rawHTML, md, warnings := processDocHTML(doc)

		dir := filepath.Dir(path)

		intermediateDoc := xhtml.OuterHTML(intDoc)
		richTextStr := xhtml.OuterHTML(richText)
		rawHTMLStr := xhtml.OuterHTML(rawHTML)

		testfile.Equalish(t, filepath.Join(dir, "intermediate.html"), intermediateDoc)
		testfile.Equalish(t, filepath.Join(dir, "rich.html"), richTextStr)
		testfile.Equalish(t, filepath.Join(dir, "raw.html"), rawHTMLStr)
		testfile.Equalish(t, filepath.Join(dir, "article.md"), md)
		testfile.EqualJSON(t, filepath.Join(dir, "metadata.json"), metadata)
		testfile.EqualJSON(t, filepath.Join(dir, "embeds.json"), embeds)
		testfile.EqualJSON(t, filepath.Join(dir, "warnings.json"), warnings)
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
