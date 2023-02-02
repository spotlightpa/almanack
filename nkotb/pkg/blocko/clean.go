package blocko

import (
	"strings"

	"github.com/spotlightpa/nkotb/pkg/xhtml"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Clean(root *html.Node) {
	mergeSiblings(root)
	removeEmptyP(root)
	replaceWhitespace(root)
}

func mergeSiblings(root *html.Node) {
	// find all matches first
	var inlineSiblings []*html.Node
	xhtml.VisitAll(root, func(n *html.Node) {
		brother := n.NextSibling
		if brother == nil ||
			!xhtml.InlineElements[n.DataAtom] ||
			n.DataAtom != brother.DataAtom ||
			!slices.Equal(n.Attr, brother.Attr) {
			return
		}
		inlineSiblings = append(inlineSiblings, n)
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
	var emptyP []*html.Node
	xhtml.VisitAll(root, func(n *html.Node) {
		if n.DataAtom == atom.P && xhtml.IsEmpty(n) {
			emptyP = append(emptyP, n)
			return
		}
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
		if n.Type == html.TextNode {
			n.Data = whitespaceReplacer.Replace(n.Data)
		}
	})
}
