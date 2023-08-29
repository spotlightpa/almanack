package blocko

import (
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestIsEmpty(t *testing.T) {
	tcases := map[string]struct {
		in    string
		empty bool
	}{
		"span":       {"<span></span>", true},
		"div":        {"<div></div>", false},
		"span-space": {"<span> </span>", true},
		"span-nl":    {"<span>\n\n</span>", true},
		"text-blank": {"<span>\n</span> ", true},
		"text":       {"x", false},
		"span-text":  {"<span></span>x", false},
	}
	for name, tc := range tcases {
		t.Run(name, func(t *testing.T) {
			p := &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.P,
				Data:     "p",
			}
			children, err := html.ParseFragment(strings.NewReader(tc.in), p)
			be.NilErr(t, err)
			for _, c := range children {
				p.AppendChild(c)
			}
			be.DebugLog(t, "got: %q", xhtml.OuterHTML(p))
			be.Equal(t, isEmpty(p), tc.empty)
		})
	}
}
