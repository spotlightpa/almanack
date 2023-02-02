package xhtml

import "golang.org/x/net/html"

func New(tag string, attrs ...string) *html.Node {
	var attrslice []html.Attribute
	if len(attrs) > 0 {
		if len(attrs)%2 != 0 {
			panic("uneven number of attr/value pairs")
		}
		attrslice = make([]html.Attribute, len(attrs)/2)
		for i := range attrslice {
			attrslice[i] = html.Attribute{
				Key: attrs[i*2],
				Val: attrs[i*2+1],
			}
		}
	}
	return &html.Node{
		Type: html.ElementNode,
		Data: tag,
		Attr: attrslice,
	}
}

func LastChildOrNew(p *html.Node, tag string, attrs ...string) *html.Node {
	if p.LastChild != nil && p.LastChild.Data == tag {
		return p.LastChild
	}
	n := New(tag, attrs...)
	p.AppendChild(n)
	return n
}

func AppendText(n *html.Node, text string) {
	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})
}
