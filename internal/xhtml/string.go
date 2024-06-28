package xhtml

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// ToBuffer returns a *bytes.Buffer containing the outerHTML of n.
func ToBuffer(n *html.Node) *bytes.Buffer {
	var buf bytes.Buffer
	buf.Grow(4 * 1 >> 10)
	if err := html.Render(&buf, n); err != nil {
		panic(err)
	}
	return &buf
}

// OuterHTML returns a serialized node.
func OuterHTML(n *html.Node) string {
	var buf strings.Builder
	buf.Grow(256)
	if err := html.Render(&buf, n); err != nil {
		panic(err)
	}
	return buf.String()
}

// InnerHTML returns the serialized markup contained within n.
func InnerHTML(n *html.Node) string {
	var buf strings.Builder
	buf.Grow(256)

	for c := range ChildNodes(n) {
		if err := html.Render(&buf, c); err != nil {
			panic(err)
		}
	}
	return buf.String()
}

// InnerHTMLBlocks is the same as InnerHTML,
// but it separates top level nodes with a line break.
func InnerHTMLBlocks(n *html.Node) string {
	var buf strings.Builder
	buf.Grow(256)

	for c := range ChildNodes(n) {
		if err := html.Render(&buf, c); err != nil {
			panic(err)
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// TextContent joins and trims the text node children of n.
func TextContent(n *html.Node) string {
	var buf strings.Builder
	buf.Grow(256)

	for n := range All(n) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
	}

	return strings.TrimSpace(buf.String())
}
