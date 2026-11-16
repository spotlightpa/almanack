package blocko

import (
	"github.com/spotlightpa/almanack/internal/utils/stringx"
	"golang.org/x/exp/collection/set"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var BlockElements = set.Of(
	atom.P,
	atom.H1,
	atom.H2,
	atom.H3,
	atom.H4,
	atom.H5,
	atom.H6,
	atom.Ul,
	atom.Ol,
)

var InlineElements = set.Of(
	atom.A,
	atom.Abbr,
	atom.Acronym,
	atom.B,
	atom.Bdi,
	atom.Bdo,
	atom.Big,
	atom.Cite,
	atom.Code,
	atom.Del,
	atom.Dfn,
	atom.Em,
	atom.I,
	atom.Ins,
	atom.Kbd,
	atom.Label,
	atom.Mark,
	atom.Meter,
	atom.Output,
	atom.Q,
	atom.Ruby,
	atom.S,
	atom.Samp,
	atom.Small,
	atom.Span,
	atom.Strong,
	atom.Sub,
	atom.Sup,
	atom.U,
	atom.Tt,
	atom.Var,
	atom.Wbr,
)

func isEmpty(n *html.Node) bool {
	for n := range n.Descendants() {
		switch n.Type {
		case html.TextNode:
			if !stringx.IsWhitespace(n.Data) {
				return false
			}
		case html.ElementNode:
			if !InlineElements.Contains(n.DataAtom) {
				return false
			}
		default:
			return false
		}
	}
	return true
}
