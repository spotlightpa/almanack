package xhtml

import (
	"iter"
	"strings"

	"github.com/spotlightpa/almanack/internal/iterx"
	"golang.org/x/net/html"
)

// Children returns a seq of the children of n.
func Children(n *html.Node) iter.Seq[*html.Node] {
	return func(yield func(*html.Node) bool) {
		if n == nil {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if !yield(c) {
				return
			}
		}
	}
}

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

// UnnestChildren has all of the children of node adopted by its parent,
// and then it removes the node.
func UnnestChildren(n *html.Node) {
	if n == nil {
		return
	}
	if n.Parent == nil {
		return
	}
	children := iterx.Collect(Children(n))
	RemoveAll(children)
	for _, c := range children {
		n.Parent.InsertBefore(c, n)
	}
	n.Parent.RemoveChild(n)
}
