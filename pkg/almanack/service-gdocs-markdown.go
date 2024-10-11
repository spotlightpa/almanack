package almanack

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/lazy"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func intermediateDocToMarkdown(doc *html.Node) string {
	// Remove partner exclusive text
	for dataEl := range dataEls(doc, dtPartnerText) {
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
			container.InsertBefore(&html.Node{
				Type: html.RawNode,
				Data: "{{<toc>}}",
			}, container.FirstChild)
			container.AppendChild(&html.Node{
				Type: html.RawNode,
				Data: "{{</toc>}}",
			})
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

// Fundraise Up embeds
var fruRe = lazy.RE(`^\s*<a href="#(\w+)" style="display: none"></a>\s*$`)

func replaceSpotlightShortcodes(s string) string {
	if s == "" {
		return s
	}
	if strings.Contains(s, "{{<") && strings.Contains(s, ">}}") {
		return s
	}
	if matches := fruRe().FindStringSubmatch(s); len(matches) == 2 {
		return fmt.Sprintf(`{{<fundraiseup id="%s">}}`, matches[1])
	}
	n, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return s
	}
	var buf strings.Builder
	isFirst := true

	// $("div[data-spl-embed-version=1]")
	els := xhtml.SelectAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Div && xhtml.Attr(n, "data-spl-embed-version") == "1"
	})
	for el := range els {
		if !isFirst {
			buf.WriteString("\n")
		}
		isFirst = false
		netloc := xhtml.Attr(el, "data-spl-src")
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

	// $("iframe[src~=vimeo]")
	els = xhtml.SelectAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Iframe &&
			strings.HasPrefix(xhtml.Attr(n, "src"), "https://player.vimeo.com/video/")
	})
	for el := range els {
		if !isFirst {
			buf.WriteString("\n")
		}
		isFirst = false
		src := xhtml.Attr(el, "src")
		u, err := url.Parse(src)
		if err != nil {
			continue
		}
		secret := u.Query().Get("h")
		id := strings.TrimPrefix(u.Path, "/video/")

		buf.WriteString("{{<vimeo id=\"")
		buf.WriteString(escapeAttr(id))
		if secret != "" {
			buf.WriteString("\" secret=\"")
			buf.WriteString(escapeAttr(secret))
		}
		buf.WriteString("\" >}}")
	}

	// $("div.flourish-embed.flourish-table")
	els = xhtml.SelectAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Div &&
			strings.Contains(xhtml.Attr(n, "class"), "flourish-embed")
	})
	for el := range els {
		if !isFirst {
			buf.WriteString("\n")
		}
		isFirst = false
		src := xhtml.Attr(el, "data-src")
		buf.WriteString(`{{<flourish src="`)
		buf.WriteString(escapeAttr(src))
		buf.WriteString(`" >}}`)
	}

	// $("iframe[src~=datawrapper.dwcdn.net]")
	els = xhtml.SelectAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Iframe &&
			strings.Contains(xhtml.Attr(n, "src"), "datawrapper.dwcdn.net")
	})
	for el := range els {
		if !isFirst {
			buf.WriteString("\n")
		}
		isFirst = false
		src := xhtml.Attr(el, "src")
		height := xhtml.Attr(el, "height")
		buf.WriteString(`{{<datawrapper src="`)
		buf.WriteString(escapeAttr(src))
		buf.WriteString(`" height="`)
		buf.WriteString(escapeAttr(height))
		buf.WriteString(`" >}}`)
	}

	// $("iframe[src~=https://www.scribd.com/embeds/]")
	els = xhtml.SelectAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Iframe &&
			strings.HasPrefix(xhtml.Attr(n, "src"), "https://www.scribd.com/embeds/")
	})
	for el := range els {
		if !isFirst {
			buf.WriteString("\n")
		}
		isFirst = false
		src := xhtml.Attr(el, "src")
		buf.WriteString(`{{<scribd src="`)
		buf.WriteString(escapeAttr(src))
		buf.WriteString(`" >}}`)
	}

	// $("div[data-tf-live]")
	els = xhtml.SelectAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Div &&
			xhtml.Attr(n, "data-tf-live") != ""
	})
	for el := range els {
		if !isFirst {
			buf.WriteString("\n")
		}
		isFirst = false
		id := xhtml.Attr(el, "data-tf-live")
		buf.WriteString(`{{<typeform id="`)
		buf.WriteString(escapeAttr(id))
		buf.WriteString(`" >}}`)
	}

	// $("iframe[src^=https://youtube.com/embeds/]")
	els = xhtml.SelectAll(n, func(n *html.Node) bool {
		if n.DataAtom != atom.Iframe {
			return false
		}
		src := xhtml.Attr(n, "src")
		return strings.HasPrefix(src, "https://www.youtube.com/embed/") ||
			strings.HasPrefix(src, "https://www.youtube-nocookie.com/embed/")
	})
	for el := range els {
		if !isFirst {
			buf.WriteString("\n")
		}
		isFirst = false
		src := xhtml.Attr(el, "src")
		u := must.Get(url.Parse(src))
		id := strings.TrimPrefix(u.Path, "/embed/")
		buf.WriteString(`{{<youtube id="`)
		buf.WriteString(escapeAttr(id))
		if start := u.Query().Get("start"); start != "" {
			buf.WriteString(`" start="`)
			buf.WriteString(escapeAttr(start))
		}
		buf.WriteString(`" loading="lazy">}}`)
	}

	if buf.Len() > 0 {
		return buf.String()
	}
	// Unknown embed type
	attr := escapeAttr(s)
	return fmt.Sprintf(`{{<embed/raw srcdoc="%s">}}`, attr)
}

func escapeAttr(s string) string {
	attr := html.EscapeString(s)
	attr = strings.ReplaceAll(attr, "\n", "&#10;")
	return attr
}
