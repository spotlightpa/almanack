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

func outputBlocks(w io.Writer, root *html.Node, depth int) (err error) {
	for p := root.FirstChild; p != nil; p = p.NextSibling {
		if depth > 0 {
			switch {
			case p.Parent.Parent.DataAtom == atom.Ol &&
				p == root.FirstChild:
				fmt.Fprint(w, "1. ")
			case p == root.FirstChild:
				fmt.Fprint(w, "- ")
			default:
				fmt.Fprint(w, strings.Repeat("    ", depth))
			}
		}
		block := strings.TrimSpace(blockToString(p))
		if _, err = io.WriteString(w, block); err != nil {
			return err
		}
		if _, err = io.WriteString(w, "\n\n"); err != nil {
			return err
		}
	}
	return nil
}

func blockToString(p *html.Node) string {
	if !xhtml.MarkdownBlockElements[p.DataAtom] {
		return xhtml.ToString(p)
	}
	prefix := ""
	switch p.DataAtom {
	case atom.Ul, atom.Ol:
		var buf strings.Builder
		for c := p.FirstChild; c != nil; c = c.NextSibling {
			outputBlocks(&buf, c, 1)
		}
		return buf.String()
	case atom.H1:
		prefix = "# "
	case atom.H2:
		prefix = "## "
	case atom.H3:
		prefix = "### "
	case atom.H4:
		prefix = "#### "
	case atom.H5:
		prefix = "##### "
	case atom.H6:
		prefix = "###### "
	}
	contents := xhtml.ContentsToString(p)
	contents = strings.TrimSpace(contents)
	contents = replaceSpecial.Replace(contents)
	return prefix + contents
}

var replaceSpecial = strings.NewReplacer(
	"\u2028", "<br />",
	" ", "&nbsp;",
)
