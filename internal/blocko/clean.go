package blocko

import (
	"slices"
	"strings"

	"github.com/earthboundkid/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Clean(root *html.Node) {
	MergeSiblings(root)
	RemoveEmptyP(root)
	fixBareLI(root)
	replaceWhitespace(root)
	replaceSpecials(root)
}

func MergeSiblings(root *html.Node) {
	// find all matches first
	inlineSiblings := xhtml.SelectSlice(root, func(n *html.Node) bool {
		brother := n.NextSibling
		return brother != nil &&
			inlineElements[n.DataAtom] &&
			n.DataAtom == brother.DataAtom &&
			slices.Equal(n.Attr, brother.Attr)
	})
	// then do mutation.
	// no mutating while iterating!
	// go in reverse order
	// in case there are several siblings to merge
	for _, n := range slices.Backward(inlineSiblings) {
		xhtml.AdoptChildren(n, n.NextSibling)
		n.Parent.RemoveChild(n.NextSibling)
	}
}

func RemoveEmptyP(root *html.Node) {
	emptyP := xhtml.SelectSlice(root, func(n *html.Node) bool {
		return n.DataAtom == atom.P && isEmpty(n)
	})
	for _, n := range emptyP {
		n.Parent.RemoveChild(n)
	}
}

func RemoveMarks(root *html.Node) {
	marks := xhtml.SelectSlice(root, xhtml.WithAtom(atom.Mark))
	for _, mark := range marks {
		xhtml.UnnestChildren(mark)
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
	for n := range root.Descendants() {
		if n.Type != html.TextNode {
			continue
		}
		// Ignore children of pre/code
		codeblock := xhtml.Closest(n, func(n *html.Node) bool {
			return n.DataAtom == atom.Pre ||
				n.DataAtom == atom.Code ||
				n.Type == html.RawNode
		})
		if codeblock == nil {
			n.Data = whitespaceReplacer.Replace(n.Data)
		}
	}
}

var specialReplacer = strings.NewReplacer(
	`\`, `\\`,
	`#`, `\#`,
	`*`, `\*`,
	`+`, `\+`,
	`[`, `\[`,
	`]`, `\]`,
	`^`, `\^`,
	`_`, `\_`,
	`~`, `\~`,
	"`", "\\`",
)

func replaceSpecials(root *html.Node) {
	for n := range root.Descendants() {
		if n.Type != html.TextNode {
			continue
		}
		// Ignore children not of p
		codeblock := xhtml.Closest(n, xhtml.WithAtom(atom.P))
		if codeblock == nil {
			continue
		}
		n.Data = specialReplacer.Replace(n.Data)
	}
}

func fixBareLI(root *html.Node) {
	bareLIs := xhtml.SelectSlice(root, func(n *html.Node) bool {
		child := n.FirstChild
		return n.DataAtom == atom.Li && child != nil &&
			(child.Type == html.TextNode ||
				inlineElements[child.DataAtom])
	})
	for _, li := range bareLIs {
		p := xhtml.New("p")
		xhtml.AdoptChildren(p, li)
		li.AppendChild(p)
	}
}
