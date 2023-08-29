// Package xhtml makes x/net/html easier
package xhtml

import (
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

// Find returns the first matching child node or nil.
func Find(n *html.Node, match func(*html.Node) bool) *html.Node {
	var found *html.Node
	all(n, func(n *html.Node) bool {
		if match(n) {
			found = n
			return _break
		}
		return _continue
	})
	return found
}

// VisitAll vists child nodes in depth-first pre-order.
func VisitAll(n *html.Node, callback func(*html.Node)) {
	all(n, func(n *html.Node) bool {
		callback(n)
		return _continue
	})
}

// FindAll returns a slice of matching nodes.
func FindAll(root *html.Node, match func(*html.Node) bool) []*html.Node {
	var found []*html.Node
	VisitAll(root, func(n *html.Node) {
		if match(n) {
			found = append(found, n)
		}
	})
	return found
}

// Closest traverses the node and its parents until it finds a node that matches.
func Closest(n *html.Node, match func(*html.Node) bool) *html.Node {
	for n != nil {
		if match(n) {
			return n
		}
		n = n.Parent
	}
	return nil
}

func WithAtom(a atom.Atom) func(n *html.Node) bool {
	return func(n *html.Node) bool {
		return n.DataAtom == a
	}
}

var WithBody = WithAtom(atom.Body)
