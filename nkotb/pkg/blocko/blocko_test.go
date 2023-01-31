package blocko_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/nkotb/pkg/blocko"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func toString(n *html.Node) string {
	var buf strings.Builder
	html.Render(&buf, n)
	return buf.String()
}

func read(t testing.TB, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
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
			be.NilErr(t, err)
			for _, c := range children {
				p.AppendChild(c)
			}
			be.DebugLog(t, "got: %q", toString(p))
			be.Equal(t, blocko.IsEmpty(p), tc.empty)
		})
	}
}

func TestGoldenFiles(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*.html")
	be.NilErr(t, err)
	for i := range inputs {
		inHTML := inputs[i]
		name := filepath.Base(inHTML)
		name = strings.TrimSuffix(name, ".html")
		t.Run(name, func(t *testing.T) {
			in := strings.NewReader(read(t, inHTML))

			var buf strings.Builder
			blocko.HTMLToMarkdown(&buf, in)

			wantMD := strings.TrimSuffix(inHTML, ".html") + ".md"
			want := read(t, wantMD)
			be.NilErr(t, err)
			got := buf.String()
			be.Debug(t, func() {
				bad := filepath.Join("testdata", name+".bad.md")
				os.WriteFile(bad, []byte(got), 0644)
			})

			be.Equal(t, want, got)
		})
	}
}
