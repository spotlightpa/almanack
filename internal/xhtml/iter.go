// Package xhtml makes x/net/html easier
package xhtml

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	done = iota
	continue_
)

func breadthFirst(n *html.Node, yield func(*html.Node) int8) {
	stack := make([]*html.Node, 1, 10)
	stack[0] = n
	for len(stack) > 0 {
		// Pop head of the stack
		var head *html.Node
		head, stack = stack[0], stack[1:]

		if yield(head) == done {
			return
		}

		// Add the head node's children to the stack then loop
		for c := head.FirstChild; c != nil; c = c.NextSibling {
			stack = append(stack, c)
		}
	}
}

func Find(n *html.Node, callback func(*html.Node) bool) *html.Node {
	var found *html.Node
	breadthFirst(n, func(n *html.Node) int8 {
		if callback(n) {
			found = n
			return done
		}
		return continue_
	})
	return found
}

func VisitAll(n *html.Node, callback func(*html.Node)) {
	breadthFirst(n, func(n *html.Node) int8 {
		callback(n)
		return continue_
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
