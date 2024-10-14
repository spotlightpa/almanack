// Package xhtml makes x/net/html easier
package xhtml

import (
	"iter"
	"slices"

	"github.com/spotlightpa/almanack/internal/iterx"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func all(n *html.Node, yield func(*html.Node) bool) bool {
	for c := range ChildNodes(n) {
		if !yield(c) || !all(c, yield) {
			return false
		}
	}
	return true
}

// All vists all child nodes in depth-first pre-order.
func All(n *html.Node) iter.Seq[*html.Node] {
	return func(yield func(*html.Node) bool) {
		all(n, yield)
	}
}

// SelectAll returns an iterator yielding matching nodes.
func SelectAll(n *html.Node, match func(*html.Node) bool) iter.Seq[*html.Node] {
	return iterx.Filter(All(n), match)
}

// SelectSlice returns a slice of child nodes matched by the selector.
func SelectSlice(n *html.Node, match func(*html.Node) bool) []*html.Node {
	return slices.Collect(SelectAll(n, match))
}

// Select returns the first child node matched by the selector or nil.
func Select(n *html.Node, match func(*html.Node) bool) *html.Node {
	return iterx.First(SelectAll(n, match))
}

// Parents returns an iterator traversing the node's parents.
func Parents(n *html.Node) iter.Seq[*html.Node] {
	return func(yield func(*html.Node) bool) {
		for p := n.Parent; p != nil && yield(p); p = p.Parent {
		}
	}
}

// Closest traverses the node and its parents until it finds a node that matches.
func Closest(n *html.Node, match func(*html.Node) bool) *html.Node {
	if match(n) {
		return n
	}
	for p := range Parents(n) {
		if match(p) {
			return p
		}
	}
	return nil
}

func WithAtom(a atom.Atom) func(n *html.Node) bool {
	return func(n *html.Node) bool {
		return n.DataAtom == a
	}
}

var WithBody = WithAtom(atom.Body)
