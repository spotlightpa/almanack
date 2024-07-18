package almanack

import (
	"fmt"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/iterx"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func fixRawHTMLPlaceholders(doc *html.Node) {
	// Remove Spotlight PA exclusives
	for dataEl, _ := range iterx.Concat2(
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
	for dataEl, value := range dataEls(doc, dtDBEmbed) {
		dbembed := dbEmbedFromString(value)
		switch dbembed.Type {
		// Replace images with red placeholder text
		case db.ImageEmbedTag:
			placeholder := xhtml.New("h2", "style", "color: red;")
			xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", dbembed.N))
			xhtml.ReplaceWith(dataEl, placeholder)

		// Include other embeds as is
		case db.RawEmbedTag, db.ToCEmbedTag, db.PartnerRawEmbedTag:
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: dbembed.Value.(string),
			})
		default:
			panic("unknown embed type: " + dbembed.Type)
		}
	}
	if el := xhtml.Select(doc, xhtml.WithAtom(atom.Data)); el != nil {
		panic("unprocessed data element: " + xhtml.OuterHTML(el))
	}
}
