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

	Clean(bNode)

	return outputBlocks(w, bNode, 0)
}

func outputBlocks(w io.Writer, bNode *html.Node, depth int) (err error) {
	var wbuf strings.Builder
	first := true
loop:
	for p := bNode.FirstChild; p != nil; p = p.NextSibling {
		if first {
			first = false
		} else {
			fmt.Fprint(w, "\n\n")
		}
		if depth > 0 {
			switch {
			case p.Parent.Parent.DataAtom == atom.Ol &&
				p == bNode.FirstChild:
				fmt.Fprint(w, "1. ")
			case p == bNode.FirstChild:
				fmt.Fprint(w, "- ")
			default:
				fmt.Fprint(w, strings.Repeat("    ", depth))
			}
		}
		if !xhtml.BlockElements[p.DataAtom] {
			if err = html.Render(&wbuf, p); err != nil {
				return err
			}
			output(w, &wbuf)
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
		output(w, &wbuf)
	}
	return nil
}

func output(w io.Writer, wbuf *strings.Builder) {
	s := wbuf.String()
	s = strings.TrimSpace(s)
	wbuf.Reset()
	fmt.Fprint(w, s)
}
