package blocko

import (
	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"strings"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	be := assert.FailNow(t)
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
		"nested":     {"<a><b>\n</b></a> ", true},
		"nested-x":   {"<a><b>x</b></a> ", false},
	}
	for name, tc := range tcases {
		t.Run(name, func(t *testing.T) {
			p := &html.Node{
				Type:     html.ElementNode,
				DataAtom: atom.P,
				Data:     "p",
			}
			children, err := html.ParseFragment(strings.NewReader(tc.in), p)
			be.Zero(err)
			for _, c := range children {
				p.AppendChild(c)
			}
			t.Logf("got: %q", xhtml.OuterHTML(p))
			be.Equal(tc.empty, isEmpty(p))
		})
	}
}
