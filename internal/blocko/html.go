//go:build goexperiment.rangefunc

package blocko

import (
	"strings"

	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var markdownBlockElements = map[atom.Atom]bool{
	atom.P:  true,
	atom.H1: true,
	atom.H2: true,
	atom.H3: true,
	atom.H4: true,
	atom.H5: true,
	atom.H6: true,
	atom.Ul: true,
	atom.Ol: true,
}

var inlineElements = map[atom.Atom]bool{
	atom.A:       true,
	atom.Abbr:    true,
	atom.Acronym: true,
	atom.B:       true,
	atom.Bdi:     true,
	atom.Bdo:     true,
	atom.Big:     true,
	atom.Cite:    true,
	atom.Code:    true,
	atom.Del:     true,
	atom.Dfn:     true,
	atom.Em:      true,
	atom.I:       true,
	atom.Ins:     true,
	atom.Kbd:     true,
	atom.Label:   true,
	atom.Mark:    true,
	atom.Meter:   true,
	atom.Output:  true,
	atom.Q:       true,
	atom.Ruby:    true,
	atom.S:       true,
	atom.Samp:    true,
	atom.Small:   true,
	atom.Span:    true,
	atom.Strong:  true,
	atom.Sub:     true,
	atom.Sup:     true,
	atom.U:       true,
	atom.Tt:      true,
	atom.Var:     true,
	atom.Wbr:     true,
}

func isEmpty(n *html.Node) bool {
	root := n
	for n := range xhtml.All(n) {
		if n == root {
			continue
		}
		switch n.Type {
		case html.TextNode:
			s := strings.ReplaceAll(n.Data, "\n", " ")
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
		case html.ElementNode:
			if inlineElements[n.DataAtom] {
				continue
			}
		}
		return false
	}
	return true
}
