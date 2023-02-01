// Package xhtml makes x/net/html easier
package xhtml

import "golang.org/x/net/html"

const (
	Done     = false
	Continue = true
)

func BreadFirst(n *html.Node, yield func(*html.Node) bool) bool {
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

func Find(n *html.Node, callback func(*html.Node) *html.Node) *html.Node {
	var found *html.Node
	BreadFirst(n, func(n *html.Node) bool {
		if child := callback(n); child != nil {
			found = child
			return Done
		}
		return Continue
	})
	return found
}

func VisitAll(n *html.Node, callback func(*html.Node)) {
	BreadFirst(n, func(n *html.Node) bool {
		callback(n)
		return true
	})
}
