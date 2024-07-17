package almanack

import (
	"fmt"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func fixRawHTMLPlaceholders(rawHTML *html.Node) {
	embeds := xhtml.SelectSlice(rawHTML, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractDataTag(dataEl)
		switch embed.Type {
		default:
			panic("unknown embed type: " + embed.Type)
		case dtSpotlightRaw, dtSpotlightText:
			dataEl.Parent.RemoveChild(dataEl)
		case dtPartnerText:
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: embed.Value,
			})
		case dtDBEmbed:
			dbembed := extractDBEmbed(embed)
			switch dbembed.Type {
			case db.RawEmbedTag, db.ToCEmbedTag, db.PartnerRawEmbedTag:
				xhtml.ReplaceWith(dataEl, &html.Node{
					Type: html.RawNode,
					Data: dbembed.Value.(string),
				})
			case db.ImageEmbedTag:
				placeholder := xhtml.New("h2", "style", "color: red;")
				xhtml.AppendText(placeholder, fmt.Sprintf("Embed #%d", dbembed.N))
				xhtml.ReplaceWith(dataEl, placeholder)
			default:
				panic("unknown embed type: " + dbembed.Type)
			}
		}
	}
}
