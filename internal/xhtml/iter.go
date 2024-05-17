//go:build goexperiment.rangefunc

// Package xhtml makes x/net/html easier
package xhtml

import (
	"iter"

	"github.com/spotlightpa/almanack/internal/iterx"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	_continue = true
	_break    = false
)

func all(n *html.Node, yield func(*html.Node) bool) bool {
	if n == nil {
		return _continue
	}
	if !yield(n) {
		return _break
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if !all(c, yield) {
			return _break
		}
	}
	return _continue
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
	return iterx.Collect(SelectAll(n, match))
}

// SelectSlice returns the first child node matched by the selector or nil.
func Select(n *html.Node, match func(*html.Node) bool) *html.Node {
	return iterx.First(SelectAll(n, match))
}

// Parents returns an iterator traversing the node and its parents.
func Parents(n *html.Node) iter.Seq[*html.Node] {
	return func(yield func(*html.Node) bool) {
		for p := n; p != nil; p = p.Parent {
			if !yield(p) {
				return
			}
		}
	}
}

// Closest traverses the node and its parents until it finds a node that matches.
func Closest(n *html.Node, match func(*html.Node) bool) *html.Node {
	return iterx.First(iterx.Filter(Parents(n), match))
}

func WithAtom(a atom.Atom) func(n *html.Node) bool {
	return func(n *html.Node) bool {
		return n.DataAtom == a
	}
}

var WithBody = WithAtom(atom.Body)
