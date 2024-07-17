package almanack

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func fixMarkdownPlaceholders(rawHTML *html.Node) {
	embeds := xhtml.SelectSlice(rawHTML, xhtml.WithAtom(atom.Data))
	for _, dataEl := range embeds {
		embed := extractDataTag(dataEl)
		switch embed.Type {
		default:
			panic("unknown embed type: " + embed.Type)
		case dtPartnerText:
			dataEl.Parent.RemoveChild(dataEl)
		case dtSpotlightText:
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: embed.Value,
			})
		case dtSpotlightRaw:
			data := replaceSpotlightShortcodes(embed.Value)
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: data,
			})
		case dtDBEmbed:
			dbembed := extractDBEmbed(embed)
			switch dbembed.Type {
			case db.RawEmbedTag:
				data := replaceSpotlightShortcodes(dbembed.Value.(string))
				xhtml.ReplaceWith(dataEl, &html.Node{
					Type: html.RawNode,
					Data: data,
				})
			case db.PartnerRawEmbedTag:
				dataEl.Parent.RemoveChild(dataEl)
			case db.ToCEmbedTag:
				container := xhtml.New("div")
				must.Do(xhtml.SetInnerHTML(container, dbembed.Value.(string)))
				xhtml.ReplaceWith(dataEl, container)
				xhtml.UnnestChildren(container)
			case db.ImageEmbedTag:
				image := dbembed.Value.(db.EmbedImage)
				var widthHeight string
				if image.Width != 0 {
					widthHeight = fmt.Sprintf(`width-ratio="%d" height-ratio="%d" `,
						image.Width, image.Height,
					)
				}
				data := fmt.Sprintf(
					`{{<picture src="%s" %sdescription="%s" caption="%s" credit="%s">}}`,
					image.Path,
					widthHeight,
					html.EscapeString(strings.TrimSpace(image.Description)),
					html.EscapeString(strings.TrimSpace(image.Caption)),
					html.EscapeString(strings.TrimSpace(image.Credit)),
				)
				xhtml.ReplaceWith(dataEl, &html.Node{
					Type: html.RawNode,
					Data: data,
				})
			default:
				panic("unknown embed type: " + dbembed.Type)
			}
		}
	}
}

func replaceSpotlightShortcodes(s string) string {
	if s == "" {
		return s
	}
	if strings.Contains(s, "{{<") && strings.Contains(s, ">}}") {
		return s
	}
	n, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	// $("div[data-spl-embed-version=1]")
	divs := xhtml.SelectSlice(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Div && xhtml.Attr(n, "data-spl-embed-version") == "1"
	})
	// Unknown embed type
	if len(divs) < 1 {
		attr := escapeAttr(s)
		return fmt.Sprintf(`{{<embed/raw srcdoc="%s">}}`, attr)
	}
	var buf strings.Builder
	for i, div := range divs {
		if i != 0 {
			buf.WriteString("\n")
		}
		netloc := xhtml.Attr(div, "data-spl-src")
		u, err := url.Parse(netloc)
		if err != nil {
			return s
		}
		tag := strings.Trim(u.Path, "/")
		if !slices.Contains([]string{
			"embeds/cta",
			"embeds/donate",
			"embeds/newsletter",
			"embeds/tips",
		}, tag) {
			return s
		}
		tag = strings.TrimPrefix(tag, "embeds/")
		q := u.Query()
		buf.WriteString("{{<embed/")
		buf.WriteString(tag)
		for _, k := range slices.Sorted(maps.Keys(q)) {
			vv := q[k]
			for _, v := range vv {
				buf.WriteString(" ")
				buf.WriteString(k)
				buf.WriteString("=\"")
				buf.WriteString(escapeAttr(v))
				buf.WriteString("\"")
			}
		}
		buf.WriteString(">}}")
	}
	return buf.String()
}

func escapeAttr(s string) string {
	attr := html.EscapeString(s)
	attr = strings.ReplaceAll(attr, "\n", "&#10;")
	return attr
}
