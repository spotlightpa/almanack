package anf

import (
	"cmp"
	"fmt"
	"net/url"
	"strings"

	"github.com/earthboundkid/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ConvertToAppleNews(htmlContent, srcURL string) (*Article, error) {
	u, err := url.Parse(srcURL)
	if err != nil {
		return nil, fmt.Errorf("ConvertToAppleNews: parsing URL: %w", err)
	}
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("ConvertToAppleNews: parsing HTML: %w", err)
	}
	for c := range doc.Descendants() {
		// Absolutize URLs
		switch c.DataAtom {
		case atom.A, atom.Link:
			if href := xhtml.Attr(c, "href"); href != "" {
				if attrURL, err := u.Parse(href); err == nil {
					xhtml.SetAttr(c, "href", attrURL.String())
				}
			}
		case atom.Img, atom.Script, atom.Iframe, atom.Audio, atom.Video, atom.Source, atom.Embed:
			if src := xhtml.Attr(c, "src"); src != "" {
				if attrURL, err := u.Parse(src); err == nil {
					xhtml.SetAttr(c, "src", attrURL.String())
				}
			}
		}

	}
	return ConvertHTMLToAppleNews(doc), nil
}

func ConvertHTMLToAppleNews(doc *html.Node) *Article {
	var c converter
	c.a = templateDoc().Clone()

	titleEl := xhtml.Select(doc, xhtml.WithAtom(atom.Title))
	c.a.Title = cmp.Or(xhtml.TextContent(titleEl), "Untitled")

	c.a.Metadata.DateCreated = nil
	c.a.Metadata.DateModified = nil
	c.a.Metadata.DatePublished = nil
	c.a.Metadata.GeneratorName = "Spotlight PA Feed2ANF"
	c.a.Metadata.GeneratorVersion = "0.0.1"
	c.a.Components = nil
	c.parseNode(doc)
	return &c.a
}

type converter struct {
	a Article
}

func (c *converter) parseNode(n *html.Node) {
	switch n.DataAtom {
	case atom.H1:
		c.addHeading(n, "introStyle", "introLayout")
	case atom.H2:
		c.addHeading(n, "introStyle", "introLayout")
	case atom.H3, atom.H4, atom.H5, atom.H6:
		c.addHeading(n, "introStyle", "introLayout")
	case atom.P:
		c.addParagraph(n)
	case atom.Blockquote:
		c.addQuote(n)
	case atom.Img:
		c.addImage(n)
	case atom.Figcaption:
		c.addCaption(n)
	case atom.Ul, atom.Ol:
		c.addList(n)
	// Container elements, process children
	default:
		for child := range n.ChildNodes() {
			c.parseNode(child)
		}
	}
}

func (c *converter) addHeading(n *html.Node, style, layout string) {
	if text := xhtml.TextContent(n); text == "" {
		return
	}
	component := TextComponent{
		Text:      xhtml.InnerHTML(n),
		Format:    "html",
		TextStyle: style,
		Layout:    layout,
		Role:      "heading",
	}

	c.a.Components = append(c.a.Components, component)
}

func (c *converter) addParagraph(n *html.Node) {
	if text := xhtml.TextContent(n); text == "" {
		return
	}
	component := TextComponent{
		Text:   xhtml.InnerHTML(n),
		Format: "html",
		Role:   "body",
		Layout: "bodyLayout",
	}

	c.a.Components = append(c.a.Components, component)
}

func (c *converter) addQuote(n *html.Node) {
	if text := xhtml.TextContent(n); text == "" {
		return
	}
	component := TextComponent{
		Text:   xhtml.InnerHTML(n),
		Format: "html",
		Role:   "quote",
		Layout: "bodyLayout",
	}
	c.a.Components = append(c.a.Components, component)
}

func (c *converter) addImage(n *html.Node) {
	src := xhtml.Attr(n, "src")
	if src == "" {
		return
	}
	// Apple doesn't understand WEBP
	if strings.HasPrefix(src, "http://www.spotlightpa.com/imgproxy/insecure") {
		if withoutExt, ok := strings.CutSuffix(src, ".webp"); ok {
			src = withoutExt + ".png"
		}
	}
	component := ImageComponent{
		URL:     src,
		Caption: xhtml.Attr(n, "alt"),
		Role:    "image",
		Layout:  "bodyLayout",
	}
	c.a.Components = append(c.a.Components, component)
}

func (c *converter) addCaption(n *html.Node) {
	if text := xhtml.TextContent(n); text == "" {
		return
	}
	component := TextComponent{
		Text:   xhtml.InnerHTML(n),
		Format: "html",
		Role:   "caption",
		Layout: "bodyLayout",
	}

	c.a.Components = append(c.a.Components, component)
}

func (c *converter) addList(n *html.Node) {
	if text := xhtml.TextContent(n); text == "" {
		return
	}
	component := TextComponent{
		Text:   xhtml.OuterHTML(n),
		Format: "html",
		Role:   "body",
		Layout: "bodyLayout",
	}

	c.a.Components = append(c.a.Components, component)
}
