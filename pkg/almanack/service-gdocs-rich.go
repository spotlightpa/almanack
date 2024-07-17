package almanack

import (
	"fmt"

	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func fixRichTextPlaceholders(richText *html.Node) {
	embeds := xhtml.SelectSlice(richText, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractDataTag(dataEl)
		switch embed.Type {
		default:
			panic("unknown embed type: " + embed.Type)
		case dtSpotlightRaw, dtSpotlightText:
			dataEl.Parent.RemoveChild(dataEl)
			continue
		case dtPartnerText:
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: embed.Value,
			})
			continue
		case dtDBEmbed:
			dbembed := extractDBEmbed(embed)
			placeholder := xhtml.New("h2", "style", "color: red;")
			xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", dbembed.N))
			xhtml.ReplaceWith(dataEl, placeholder)
		}
	}
}
