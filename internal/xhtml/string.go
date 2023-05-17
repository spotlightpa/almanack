package xhtml

import (
	"bytes"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

func ToBuffer(n *html.Node) *bytes.Buffer {
	var buf bytes.Buffer
	if err := html.Render(&buf, n); err != nil {
		panic(err)
	}
	return &buf
}

var pool sync.Pool

func poolGet() *bytes.Buffer {
	if v := pool.Get(); v != nil {
		buf := v.(*bytes.Buffer)
		buf.Reset()
		return buf
	}
	return &bytes.Buffer{}
}

func ToString(n *html.Node) string {
	buf := poolGet()
	defer pool.Put(buf)
	if err := html.Render(buf, n); err != nil {
		panic(err)
	}
	return buf.String()
}

func ContentsToString(n *html.Node) string {
	buf := poolGet()
	defer pool.Put(buf)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(buf, c); err != nil {
			panic(err)
		}
	}
	return buf.String()
}

// InnerBlocksToString is the same as ContentsToString,
// but it separates top level nodes with a line break.
func InnerBlocksToString(n *html.Node) string {
	buf := poolGet()
	defer pool.Put(buf)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if err := html.Render(buf, c); err != nil {
			panic(err)
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

// InnerText joins and trims the text node children of n.
func InnerText(n *html.Node) string {
	buf := poolGet()
	defer pool.Put(buf)

	VisitAll(n, func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		}
	})
	return strings.TrimSpace(buf.String())
}
