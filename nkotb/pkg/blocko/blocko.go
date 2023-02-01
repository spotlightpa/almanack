// Package blocko converts HTML to Markdownish text.
package blocko

import (
	"fmt"
	"io"
	"strings"

	"github.com/spotlightpa/nkotb/pkg/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func HTMLToMarkdown(w io.Writer, r io.Reader) error {
	bNode, err := prep(r)
	if err != nil {
		return err
	}

	return outputBlocks(w, bNode, 0)
}

func outputBlocks(w io.Writer, bNode *html.Node, depth int) (err error) {
	var (
		wbuf    strings.Builder
		needsNL = false
	)
loop:
	for p := bNode.FirstChild; p != nil; p = p.NextSibling {
		if needsNL {
			fmt.Fprint(w, "\n\n")
		}
		if depth > 0 {
			if p == bNode.FirstChild {
				fmt.Fprint(w, "- ")
			} else {
				fmt.Fprint(w, strings.Repeat("    ", depth))
			}
		}
		if !xhtml.BlockElements[p.DataAtom] {
			if err = html.Render(&wbuf, p); err != nil {
				return err
			}
			needsNL = output(w, &wbuf)
			continue
		}
		if xhtml.IsEmpty(p) {
			fmt.Fprint(w, "")
			needsNL = false
			continue
		}
		switch p.DataAtom {
		case atom.H1:
			wbuf.WriteString("# ")
		case atom.H2:
			wbuf.WriteString("## ")
		case atom.H3:
			wbuf.WriteString("### ")
		case atom.H4:
			wbuf.WriteString("#### ")
		case atom.H5:
			wbuf.WriteString("##### ")
		case atom.H6:
			wbuf.WriteString("###### ")
		case atom.Ul, atom.Ol:
			for c := p.FirstChild; c != nil; c = c.NextSibling {
				if err = outputBlocks(w, c, depth+1); err != nil {
					return err
				}
				fmt.Fprint(w, "\n")
			}
			continue loop
		}
		for c := p.FirstChild; c != nil; c = c.NextSibling {
			if err = html.Render(&wbuf, c); err != nil {
				return err
			}
		}
		needsNL = output(w, &wbuf)
	}
	return nil
}

func output(w io.Writer, wbuf *strings.Builder) bool {
	s := wbuf.String()
	s = strings.TrimSpace(s)
	wbuf.Reset()
	fmt.Fprint(w, s)
	return s != ""
}
