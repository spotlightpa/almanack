package xhtml_test

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
			be.DebugLog(t, "got: %q", xhtml.ToString(p))
			be.Equal(t, xhtml.IsEmpty(p), tc.empty)
		})
	}
}

func TestSetInnerHTML(t *testing.T) {
	n := xhtml.New("p")
	be.NilErr(t, xhtml.SetInnerHTML(n, "Hello, <i>World!</i>"))
	be.Equal(t, `<p>Hello, <i>World!</i></p>`, xhtml.ToString(n))

	be.NilErr(t, xhtml.SetInnerHTML(n, "Jello, <i>World!</i>"))
	be.Equal(t, `<p>Jello, <i>World!</i></p>`, xhtml.ToString(n))

	n = xhtml.New("script")
	be.NilErr(t, xhtml.SetInnerHTML(n, "let i = 1 > 2"))
	be.Equal(t, `<script>let i = 1 > 2</script>`, xhtml.ToString(n))

	n = xhtml.New("p")
	be.NilErr(t, xhtml.SetInnerHTML(n, "</a></b>"))
	be.Equal(t, `<p></p>`, xhtml.ToString(n))
}
