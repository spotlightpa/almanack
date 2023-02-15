package blocko

import (
	"strings"

	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Clean(root *html.Node) {
	mergeSiblings(root)
	removeEmptyP(root)
	replaceWhitespace(root)
	fixBareLI(root)
}

func mergeSiblings(root *html.Node) {
	// find all matches first
	inlineSiblings := xhtml.FindAll(root, func(n *html.Node) bool {
		brother := n.NextSibling
		return brother != nil &&
			xhtml.InlineElements[n.DataAtom] &&
			n.DataAtom == brother.DataAtom &&
			slices.Equal(n.Attr, brother.Attr)
	})
	// then do mutation.
	// no mutating while iterating!
	// go in reverse order
	// in case there are several siblings to merge
	for i := len(inlineSiblings) - 1; i >= 0; i-- {
		n := inlineSiblings[i]
		xhtml.AdoptChildren(n, n.NextSibling)
		n.Parent.RemoveChild(n.NextSibling)
	}
}

func removeEmptyP(root *html.Node) {
	emptyP := xhtml.FindAll(root, func(n *html.Node) bool {
		return n.DataAtom == atom.P && xhtml.IsEmpty(n)
	})
	for _, n := range emptyP {
		n.Parent.RemoveChild(n)
	}
}

var whitespaceReplacer = strings.NewReplacer(
	"\r", " ",
	"\n", " ",
	"\v", "\u2028",
	"\u2029", "\u2028",
	"  ", " ",
)

func replaceWhitespace(root *html.Node) {
	xhtml.VisitAll(root, func(n *html.Node) {
		if n.Type != html.TextNode {
			return
		}
		// Ignore children of pre/code
		codeblock := xhtml.Closest(n, func(n *html.Node) bool {
			return n.DataAtom == atom.Pre || n.DataAtom == atom.Code
		})
		if codeblock == nil {
			n.Data = whitespaceReplacer.Replace(n.Data)
		}
	})
}

func fixBareLI(root *html.Node) {
	bareLIs := xhtml.FindAll(root, func(n *html.Node) bool {
		child := n.FirstChild
		return n.DataAtom == atom.Li &&
			(child.Type == html.TextNode ||
				xhtml.InlineElements[child.DataAtom])
	})
	for _, li := range bareLIs {
		p := xhtml.New("p")
		xhtml.AdoptChildren(p, li)
		li.AppendChild(p)
	}
}
