package almanack

import (
	"fmt"

	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/iterx"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func intermediateDocToPartnerHTML(doc *html.Node) *html.Node {
	rawHTML := xhtml.Clone(doc)
	// Remove Spotlight PA exclusives
	for dataEl := range iterx.Concat2(
		dataEls(rawHTML, dtSpotlightText),
		dataEls(rawHTML, dtSpotlightRaw),
	) {
		dataEl.Parent.RemoveChild(dataEl)
	}
	// Include partner text as is
	for dataEl, text := range dataEls(rawHTML, dtPartnerText) {
		xhtml.ReplaceWith(dataEl, &html.Node{
			Type: html.RawNode,
			Data: text,
		})
	}
	for dataEl, value := range dataEls(rawHTML, dtDBEmbed) {
		dbembed := dbEmbedFromString(value)
		switch dbembed.Type {
		// Replace images with red placeholder text
		case db.ImageEmbedTag:
			if imgTag := dbembed.Value.(db.EmbedImage); imgTag.Kind == "spl" {
				dataEl.Parent.RemoveChild(dataEl)
				continue
			}
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
	if el := xhtml.Select(rawHTML, xhtml.WithAtom(atom.Data)); el != nil {
		panic("unprocessed data element: " + xhtml.OuterHTML(el))
	}
	return rawHTML
}
