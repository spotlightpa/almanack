package blocko

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spotlightpa/nkotb/pkg/xhtml"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	"golang.org/x/exp/slices"
	nethtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func prep(r io.Reader) (*nethtml.Node, error) {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)

	var out bytes.Buffer
	if err := m.Minify("text/html", &out, r); err != nil {
		return nil, err
	}
	doc, err := nethtml.Parse(&out)
	if err != nil {
		return nil, err
	}
	body := xhtml.Find(doc, func(n *nethtml.Node) *nethtml.Node {
		if n.DataAtom == atom.Body {
			return n
		}
		return nil
	})
	if body == nil {
		return nil, fmt.Errorf("could not find body")
	}

	mergeSiblings(body)

	return body, nil
}

func mergeSiblings(root *nethtml.Node) {
	// find all matches first
	var inlineSiblings []*nethtml.Node
	xhtml.VisitAll(root, func(n *nethtml.Node) {
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
