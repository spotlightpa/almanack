package blocko

import (
	"strings"
	"testing"

	"github.com/earthboundkid/assert"
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
			be := assert.FailNow(t)
			children := be.OK(html.ParseFragment(strings.NewReader(tc.in), p))
			for _, c := range children {
				p.AppendChild(c)
			}
			be.Equal(tc.empty, isEmpty(p))
		})
	}
}
