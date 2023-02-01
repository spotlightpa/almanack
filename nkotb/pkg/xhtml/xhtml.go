package xhtml

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ToString(n *html.Node) string {
	var buf strings.Builder
	html.Render(&buf, n)
	return buf.String()
}

var BlockElements = map[atom.Atom]bool{
	atom.P:  true,
	atom.H1: true,
	atom.H2: true,
	atom.H3: true,
	atom.H4: true,
	atom.H5: true,
	atom.H6: true,
	atom.Ul: true,
	atom.Ol: true,
}

var InlineElements = map[atom.Atom]bool{
	atom.A:       true,
	atom.Abbr:    true,
	atom.Acronym: true,
	atom.B:       true,
	atom.Bdi:     true,
	atom.Bdo:     true,
	atom.Big:     true,
	atom.Br:      true,
	atom.Cite:    true,
	atom.Code:    true,
	atom.Del:     true,
	atom.Dfn:     true,
	atom.Em:      true,
	atom.I:       true,
	atom.Ins:     true,
	atom.Kbd:     true,
	atom.Label:   true,
	atom.Mark:    true,
	atom.Meter:   true,
	atom.Output:  true,
	atom.Q:       true,
	atom.Ruby:    true,
	atom.S:       true,
	atom.Samp:    true,
	atom.Small:   true,
	atom.Span:    true,
	atom.Strong:  true,
	atom.Sub:     true,
	atom.Sup:     true,
	atom.U:       true,
	atom.Tt:      true,
	atom.Var:     true,
	atom.Wbr:     true,
}

func IsEmpty(n *html.Node) bool {
	root := n
	n = Find(n, func(n *html.Node) *html.Node {
		if n == root {
			return nil
		}
		switch n.Type {
		case html.TextNode:
			s := strings.ReplaceAll(n.Data, "\n", " ")
			s = strings.TrimSpace(s)
			if s == "" {
				return nil
			}
		case html.ElementNode:
			if InlineElements[n.DataAtom] {
				return nil
			}
		}
		return n
	})
	return n == nil
}

func AdoptChildren(dst, src *html.Node) {
	if dst == src {
		return
	}
	for src.FirstChild != nil {
		c := src.FirstChild
		src.RemoveChild(c)
		dst.AppendChild(c)
	}
}
