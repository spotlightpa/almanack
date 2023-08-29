package xhtml_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
)

func TestAttr(t *testing.T) {
	var n *html.Node
	be.Equal(t, "", xhtml.Attr(n, "a"))
	n = xhtml.New("span")
	be.Equal(t, `<span></span>`, xhtml.OuterHTML(n))
	be.Equal(t, "", xhtml.Attr(n, "a"))

	xhtml.SetAttr(n, "a", "b")
	be.Equal(t, `<span a="b"></span>`, xhtml.OuterHTML(n))
	be.Equal(t, "b", xhtml.Attr(n, "a"))

	xhtml.SetAttr(n, "a", "c")
	be.Equal(t, `<span a="c"></span>`, xhtml.OuterHTML(n))
	be.Equal(t, "c", xhtml.Attr(n, "a"))

	xhtml.SetAttr(n, "d", "e")
	be.Equal(t, "c", xhtml.Attr(n, "a"))
	be.Equal(t, "e", xhtml.Attr(n, "d"))
	be.Equal(t, `<span a="c" d="e"></span>`, xhtml.OuterHTML(n))

	xhtml.DeleteAttr(n, "a")
	be.Equal(t, "", xhtml.Attr(n, "a"))
	be.Equal(t, "e", xhtml.Attr(n, "d"))
	be.Equal(t, `<span d="e"></span>`, xhtml.OuterHTML(n))
}
