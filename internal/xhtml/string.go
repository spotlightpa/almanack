package xhtml

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func ToBuffer(n *html.Node) *bytes.Buffer {
	var buf bytes.Buffer
	buf.Grow(4 * 1 >> 10)
	if err := html.Render(&buf, n); err != nil {
		panic(err)
	}
	return &buf
}

func ToString(n *html.Node) string {
	var buf strings.Builder
	buf.Grow(256)
	if err := html.Render(&buf, n); err != nil {
		panic(err)
	}
	return buf.String()
}

func ContentsToString(n *html.Node) string {
	var buf strings.Builder
	buf.Grow(256)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&buf, c); err != nil {
			panic(err)
		}
	}
	return buf.String()
}

// InnerBlocksToString is the same as ContentsToString,
// but it separates top level nodes with a line break.
func InnerBlocksToString(n *html.Node) string {
	var buf strings.Builder
	buf.Grow(256)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(&buf, c); err != nil {
			panic(err)
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// InnerText joins and trims the text node children of n.
func InnerText(n *html.Node) string {
	var buf strings.Builder
	buf.Grow(256)

	VisitAll(n, func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
	})
	return strings.TrimSpace(buf.String())
}
