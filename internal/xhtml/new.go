package xhtml

import (
	"slices"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

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
		Type:     html.ElementNode,
		Data:     tag,
		DataAtom: atom.Lookup([]byte(tag)),
		Attr:     attrslice,
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

// Clone n and all of its children.
func Clone(n *html.Node) *html.Node {
	new := &html.Node{
		Type:      n.Type,
		DataAtom:  n.DataAtom,
		Data:      n.Data,
		Namespace: n.Namespace,
		Attr:      slices.Clone(n.Attr),
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		c2 := Clone(c)
		new.AppendChild(c2)
	}
	return new
}
