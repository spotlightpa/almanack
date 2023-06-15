package xhtml

import (
	"strings"

	"golang.org/x/net/html"
)

func ReplaceWith(old, new *html.Node) {
	old.Parent.InsertBefore(new, old)
	old.Parent.RemoveChild(old)
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

func AppendText(n *html.Node, text string) {
	n.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})
}

func SetInnerHTML(n *html.Node, s string) error {
	children, err := html.ParseFragment(strings.NewReader(s), n)
	if err != nil {
		return err
	}
	for c := n.FirstChild; c != nil; c = n.FirstChild {
		n.RemoveChild(c)
	}
	for _, c := range children {
		n.AppendChild(c)
	}
	return nil
}

// RemoveAll orphans the nodes it is passed.
// It ignores a node if the node is nil or already an orphan.
func RemoveAll(nodes []*html.Node) {
	for i := len(nodes) - 1; i >= 0; i-- {
		n := nodes[i]
		if n != nil && n.Parent != nil {
			n.Parent.RemoveChild(n)
		}
	}
}
