package almanack

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func intermediateDocToMarkdown(doc *html.Node) string {
	// Remove partner exclusive text
	for dataEl, _ := range dataEls(doc, dtPartnerText) {
		dataEl.Parent.RemoveChild(dataEl)
	}
	// Include Spotlight PA specific text as is
	for dataEl, value := range dataEls(doc, dtSpotlightText) {
		xhtml.ReplaceWith(dataEl, &html.Node{
			Type: html.RawNode,
			Data: value,
		})
	}
	// Replace Spotlight PA raw embeds with shortcodes if possible
	for dataEl, value := range dataEls(doc, dtSpotlightRaw) {
		data := replaceSpotlightShortcodes(value)
		xhtml.ReplaceWith(dataEl, &html.Node{
			Type: html.RawNode,
			Data: data,
		})
	}
	for dataEl, value := range dataEls(doc, dtDBEmbed) {
		dbembed := dbEmbedFromString(value)
		switch dbembed.Type {
		// Replace raw embeds with shortcodes if possible
		case db.RawEmbedTag:
			data := replaceSpotlightShortcodes(dbembed.Value.(string))
			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: data,
			})
		// Remove partner specific embeds
		case db.PartnerRawEmbedTag:
			dataEl.Parent.RemoveChild(dataEl)
		// Insert ToC
		case db.ToCEmbedTag:
			container := xhtml.New("div")
			must.Do(xhtml.SetInnerHTML(container, dbembed.Value.(string)))
			xhtml.ReplaceWith(dataEl, container)
			xhtml.UnnestChildren(container)
		// Write picture shortcode
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
	if el := xhtml.Select(doc, xhtml.WithAtom(atom.Data)); el != nil {
		panic("unprocessed data element: " + xhtml.OuterHTML(el))
	}
	return blocko.Blockize(doc)
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
