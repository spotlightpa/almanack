// Package blocko converts HTML to Markdownish text.
package blocko

import (
	"fmt"
	"io"
	"strings"

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
		if !blockElements[p.DataAtom] {
			if err = html.Render(&wbuf, p); err != nil {
				return err
			}
			needsNL = output(w, &wbuf)
			continue
		}
		if isEmpty(p) {
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

func findNode(n *html.Node, callback func(*html.Node) *html.Node) *html.Node {
	if r := callback(n); r != nil {
		return r
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := findNode(c, callback); r != nil {
			return r
		}
	}
	return nil
}

func visitAll(n *html.Node, callback func(*html.Node)) {
	findNode(n, func(n *html.Node) *html.Node {
		callback(n)
		return nil
	})
}

var blockElements = map[atom.Atom]bool{
	atom.P:  true,
	atom.H1: true,
	atom.H2: true,
	atom.H3: true,
	atom.H4: true,
	atom.H5: true,
	atom.H6: true,
	atom.Ul: true,
	atom.Ol: true,
}

var stylisticElements = map[atom.Atom]bool{
	atom.A:       true,
	atom.Abbr:    true,
	atom.Acronym: true,
	atom.B:       true,
	atom.Bdi:     true,
	atom.Bdo:     true,
	atom.Big:     true,
	atom.Br:      true,
	atom.Cite:    true,
	atom.Code:    true,
	atom.Del:     true,
	atom.Dfn:     true,
	atom.Em:      true,
	atom.I:       true,
	atom.Ins:     true,
	atom.Kbd:     true,
	atom.Label:   true,
	atom.Mark:    true,
	atom.Meter:   true,
	atom.Output:  true,
	atom.Q:       true,
	atom.Ruby:    true,
	atom.S:       true,
	atom.Samp:    true,
	atom.Small:   true,
	atom.Span:    true,
	atom.Strong:  true,
	atom.Sub:     true,
	atom.Sup:     true,
	atom.U:       true,
	atom.Tt:      true,
	atom.Var:     true,
	atom.Wbr:     true,
}

func isEmpty(n *html.Node) bool {
	root := n
	n = findNode(n, func(n *html.Node) *html.Node {
		if n == root {
			return nil
		}
		switch n.Type {
		case html.TextNode:
			s := strings.ReplaceAll(n.Data, "\n", " ")
			s = strings.TrimSpace(s)
			if s == "" {
				return nil
			}
		case html.ElementNode:
			if stylisticElements[n.DataAtom] {
				return nil
			}
		}
		return n
	})
	return n == nil
}
