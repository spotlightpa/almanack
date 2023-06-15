package xhtml

import (
	"github.com/spotlightpa/almanack/internal/slicex"
	"golang.org/x/net/html"
)

func Attr(n *html.Node, name string) string {
	if n == nil {
		return ""
	}
	for _, attr := range n.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

func SetAttr(n *html.Node, key, value string) {
	for i := range n.Attr {
		attr := &n.Attr[i]
		if attr.Key == key {
			attr.Val = value
			return
		}
	}
	n.Attr = append(n.Attr, html.Attribute{
		Key: key,
		Val: value,
	})
}

func DeleteAttr(n *html.Node, key string) {
	slicex.DeleteFunc(&n.Attr, func(a html.Attribute) bool {
		return a.Key == key
	})
}
