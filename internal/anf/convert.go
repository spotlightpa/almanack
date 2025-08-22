package anf

import (
	"bytes"
	"cmp"
	"fmt"
	"time"

	"github.com/earthboundkid/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ConvertToAppleNews(htmlContent []byte) (*Article, error) {
	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("ConvertToAppleNews: parsing HTML: %w", err)
	}
	return ConvertHTMLToAppleNews(doc), nil
}

func ConvertHTMLToAppleNews(doc *html.Node) *Article {
	var c converter
	c.a = templateDoc.Clone()

	title := extractTitle(doc)
	c.a.Title = cmp.Or(title, "Untitled")

	now := time.Now()
	c.a.Metadata.DateCreated = &now
	c.a.Metadata.DateModified = &now
	c.a.Metadata.DatePublished = &now
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
		c.addHeading(n, "default")
	case atom.H2:
		c.addHeading(n, "default")
	case atom.H3, atom.H4, atom.H5, atom.H6:
		c.addHeading(n, "default")
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

func (c *converter) addHeading(n *html.Node, style string) {
	if text := xhtml.TextContent(n); text == "" {
		return
	}
	component := TextComponent{
		Text:      xhtml.InnerHTML(n),
		Format:    "html",
		TextStyle: style,
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
	}
	c.a.Components = append(c.a.Components, component)
}

func (c *converter) addImage(n *html.Node) {
	src := xhtml.Attr(n, "src")
	if src == "" {
		return
	}
	component := ImageComponent{
		URL:     src,
		Caption: xhtml.Attr(n, "alt"),
		Role:    "image",
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
	}

	c.a.Components = append(c.a.Components, component)
}
