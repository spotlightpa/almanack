package almanack

import (
	"fmt"

	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/iterx"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func intermediateDocToPartnerRichText(doc *html.Node) {
	// Remove Spotlight PA exclusives
	for dataEl := range iterx.Concat2(
		dataEls(doc, dtSpotlightText),
		dataEls(doc, dtSpotlightRaw),
	) {
		dataEl.Parent.RemoveChild(dataEl)
	}
	// Include partner text as is
	for dataEl, text := range dataEls(doc, dtPartnerText) {
		xhtml.ReplaceWith(dataEl, &html.Node{
			Type: html.RawNode,
			Data: text,
		})
	}
	// Replace other embeds with red placeholder text
	for dataEl, value := range dataEls(doc, dtDBEmbed) {
		dbembed := dbEmbedFromString(value)
		if imgTag, ok := dbembed.Value.(db.EmbedImage); ok && imgTag.Kind == "spl" {
			dataEl.Parent.RemoveChild(dataEl)
			continue
		}
		placeholder := xhtml.New("h2", "style", "color: red;")
		xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", dbembed.N))
		xhtml.ReplaceWith(dataEl, placeholder)
	}
	if el := xhtml.Select(doc, xhtml.WithAtom(atom.Data)); el != nil {
		panic("unprocessed data element: " + xhtml.OuterHTML(el))
	}
}
