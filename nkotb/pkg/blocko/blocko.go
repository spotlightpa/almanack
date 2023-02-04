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
	var blocks []string
	for p := root.FirstChild; p != nil; p = p.NextSibling {
		blocks = append(blocks, blockToStrings(p)...)
	}
	s := strings.Join(blocks, "\n\n")
	if _, err = io.WriteString(w, s); err != nil {
		return err
	}
	return nil
}

func blockToStrings(p *html.Node) []string {
	if !xhtml.MarkdownBlockElements[p.DataAtom] {
		return []string{xhtml.ToString(p)}
	}
	prefix := ""
	switch p.DataAtom {
	case atom.Ul, atom.Ol:
		counter := 0
		if p.DataAtom == atom.Ol {
			counter = 1
		}
		var blocks []string
		for li := p.FirstChild; li != nil; li = li.NextSibling {
			marker := "- "
			if counter > 0 {
				marker = fmt.Sprintf("%d. ", counter)
				counter++
			}
			for c := li.FirstChild; c != nil; c = c.NextSibling {
				subblocks := blockToStrings(c)
				for i := range subblocks {
					if i > 0 {
						marker = "  "
					}
					subblocks[i] = marker + subblocks[i]
				}
				blocks = append(blocks, subblocks...)
				marker = "  "
			}
		}
		return blocks
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
	return []string{prefix + contents}
}

var replaceSpecial = strings.NewReplacer(
	"\u2028", "<br />",
	" ", "&nbsp;",
)
