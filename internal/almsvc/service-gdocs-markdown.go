package almsvc

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/convert/blocko"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/utils/lazy"
	"github.com/spotlightpa/almanack/internal/utils/must"
	"github.com/spotlightpa/almanack/internal/utils/shortcode"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func intermediateDocToMarkdown(doc *html.Node) string {
	doc = xhtml.Clone(doc)
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
			if image.Kind == "partner" {
				dataEl.Parent.RemoveChild(dataEl)
				continue
			}
			var tag string
			switch image.Kind {
			case "wide":
				tag = "featured/picture"
			case "left":
				tag = "featured/picture-left"
			case "right":
				tag = "featured/picture-right"
			default:
				tag = "picture"
			}

			attrs := []string{
				"src", image.Path,
			}
			if image.Width != 0 {
				attrs = append(attrs,
					"width-ratio", strconv.Itoa(image.Width),
					"height-ratio", strconv.Itoa(image.Height))
			}

			if image.Focus != "" {
				attrs = append(attrs, "focus", image.Focus)
			}
			attrs = append(attrs,
				"description", image.Description,
				"caption", image.Caption,
				"credit", image.Credit,
			)

			xhtml.ReplaceWith(dataEl, &html.Node{
				Type: html.RawNode,
				Data: shortcode.New(tag, attrs...),
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

// Digits in min-height
var digitsRe = lazy.RE(`\d+`)

func replaceSpotlightShortcodes(s string) string {
	if s == "" {
		return s
	}
	if strings.Contains(s, "{{<") && strings.Contains(s, ">}}") {
		return s
	}

	// Dewey assistant key
	if strings.Contains(s, "3a59cd0d-c005-4a65-807d-3a5e539bd515") ||
		strings.Contains(s, "ba7fd845-222f-4d30-b2d1-ef3210971999") {
		return "{{<dewey-assistant>}}"
	}

	// Fundraise Up psuedolink
	if matches := fruRe().FindStringSubmatch(s); len(matches) == 2 {
		return shortcode.New("fundraiseup", "id", matches[1])
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
		buf.WriteString(shortcode.NewMapAttrs("embed/"+tag, u.Query()))
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

		attrs := []string{"id", id}
		if secret != "" {
			attrs = append(attrs, "secret", secret)
		}
		buf.WriteString(shortcode.New("vimeo", attrs...))
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
		buf.WriteString(shortcode.New("flourish", "src", src))
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
		buf.WriteString(shortcode.New("datawrapper",
			"src", src,
			"height", height))
	}

	// $("script[src~=datawrapper.dwcdn.net]")
	els = xhtml.SelectAll(n, func(n *html.Node) bool {
		return n.DataAtom == atom.Script &&
			strings.Contains(xhtml.Attr(n, "src"), "datawrapper.dwcdn.net")
	})
	for el := range els {
		if !isFirst {
			buf.WriteString("\n")
		}
		isFirst = false
		src := xhtml.Attr(el, "src")
		// Strip the embed.js from the src because we want the iframe path
		src = strings.TrimSuffix(src, "embed.js")
		// Parent div should have style="min-height:###px".
		// We're lazy and just look for digits.
		parentStyle := xhtml.Attr(el.Parent, "style")
		height := digitsRe().FindString(parentStyle)
		buf.WriteString(shortcode.New("datawrapper",
			"src", src,
			"height", height))
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
		buf.WriteString(shortcode.New("scribd", "src", src))
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
		buf.WriteString(shortcode.New("typeform", "id", id))
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
		attrs := []string{"id", id}
		if start := u.Query().Get("start"); start != "" {
			attrs = append(attrs, "start", start)
		}
		attrs = append(attrs, "loading", "lazy")
		buf.WriteString(shortcode.New("youtube", attrs...))
	}

	if buf.Len() > 0 {
		return buf.String()
	}
	// Unknown embed type
	return shortcode.New("embed/raw", "srcdoc", s)
}
