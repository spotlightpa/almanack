// Package xhtml makes x/net/html easier
package xhtml

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	Done = iota
	Continue
)

func BreadFirst(n *html.Node, yield func(*html.Node) int8) int8 {
	if yield(n) == Done {
		return Done
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if BreadFirst(c, yield) == Done {
			return Done
		}
	}
	return Continue
}

func Find(n *html.Node, callback func(*html.Node) bool) *html.Node {
	var found *html.Node
	BreadFirst(n, func(n *html.Node) int8 {
		if callback(n) {
			found = n
			return Done
		}
		return Continue
	})
	return found
}

func VisitAll(n *html.Node, callback func(*html.Node)) {
	BreadFirst(n, func(n *html.Node) int8 {
		callback(n)
		return Continue
	})
}

func FindAll(root *html.Node, filter func(*html.Node) bool) []*html.Node {
	var found []*html.Node
	VisitAll(root, func(n *html.Node) {
		if filter(n) {
			found = append(found, n)
		}
	})
	return found
}

func Closest(n *html.Node, cond func(*html.Node) bool) *html.Node {
	for p := n.Parent; p != nil; p = p.Parent {
		if cond(p) {
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
