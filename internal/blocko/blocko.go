// Package blocko converts HTML to Markdownish text.
package blocko

import (
	"fmt"
	"strings"

	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func MinifyAndBlockize(htmlstr string) (string, error) {
	root, err := Minify(strings.NewReader(htmlstr))
	if err != nil {
		return "", err
	}

	return Blockize(root), nil
}

func Blockize(root *html.Node) string {
	Clean(root)

	var blocks []string
	for p := root.FirstChild; p != nil; p = p.NextSibling {
		blocks = append(blocks, blockToStrings(p)...)
	}

	return strings.Join(blocks, "\n\n") + "\n"
}

func blockToStrings(p *html.Node) []string {
	if !markdownBlockElements[p.DataAtom] {
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
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		if id := xhtml.Attr(p, "id"); id != "" {
			contents := xhtml.ContentsToString(p)
			contents = strings.TrimSpace(contents)
			contents = replaceSpecial.Replace(contents)
			contents = fmt.Sprintf(`<%s id="%s">%s</%s>`,
				p.Data, html.EscapeString(id), contents, p.Data,
			)
			return []string{contents}
		}
		level := int(p.Data[1] - '0')
		prefix = strings.Repeat("#", level) + " "
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
