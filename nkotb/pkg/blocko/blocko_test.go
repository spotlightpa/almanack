package blocko

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func assert(t *testing.T, v bool, format string, args ...interface{}) {
	t.Helper()
	if !v {
		t.Fatalf(format, args...)
	}
}

func assertErrNil(t *testing.T, err error) {
	t.Helper()
	assert(t, err == nil, "err != nil: %v", err)
}

func toString(n *html.Node) string {
	var buf strings.Builder
	html.Render(&buf, n)
	return buf.String()
}

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
			assertErrNil(t, err)
			for _, c := range children {
				p.AppendChild(c)
			}
			s := toString(p)
			assert(t, isEmpty(p) == tc.empty, "got: %q", s)
		})
	}
}
