package tableaux

import (
	"iter"
	"strings"

	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/utils/stringx"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Tables(root *html.Node) iter.Seq[Table] {
	return func(yield func(Table) bool) {
		tables := xhtml.SelectSlice(root, xhtml.WithAtom(atom.Table))
		for _, tblNode := range tables {
			var cells Cells
			for row := range xhtml.SelectAll(tblNode, xhtml.WithAtom(atom.Tr)) {
				tds := xhtml.SelectSlice(row, func(n *html.Node) bool {
					return n.DataAtom == atom.Td || n.DataAtom == atom.Th
				})
				cells = append(cells, tds)
			}
			if !yield(Table{tblNode, cells}) {
				return
			}
		}
	}
}

type Table struct {
	Node *html.Node
	Cells
}

func (tbl Table) RemoveFromParent() {
	tbl.Node.Parent.RemoveChild(tbl.Node)
}

func (tbl Table) ReplaceWith(n *html.Node) {
	xhtml.ReplaceWith(tbl.Node, n)
}

func (tbl Table) Set(key, value string) {
	tr := xhtml.New("tr")
	keyNode := xhtml.New("td")
	xhtml.AppendText(keyNode, key)
	tr.AppendChild(keyNode)

	valueNode := xhtml.New("td")
	xhtml.AppendText(valueNode, value)
	tr.AppendChild(valueNode)

	tbl.Node.AppendChild(tr)
	tbl.Cells = append(tbl.Cells, []*html.Node{keyNode, valueNode})
}

type Cells [][]*html.Node

func (cells Cells) At(row, col int) *html.Node {
	if row >= len(cells) {
		return &html.Node{Type: html.TextNode}
	}
	r := cells[row]
	if col >= len(r) {
		return &html.Node{Type: html.TextNode}
	}
	return r[col]
}

func slugify(n *html.Node) string {
	return strings.TrimSpace(stringx.RemoveParens(strings.ToLower(xhtml.TextContent(n))))
}

func (cells Cells) Label() string {
	return slugify(cells.At(0, 0))
}

func (cells Cells) ValueOrNext(name string) *html.Node {
	for i := range cells {
		if slugify(cells.At(i, 0)) == name {
			cell := cells.At(i, 1)
			if s := xhtml.TextContent(cell); s == "" {
				cell = cells.At(i+1, 0)
			}
			if stringx.RemoveAllWhitespace(slugify(cell)) == "n/a" {
				return &html.Node{
					Type: html.CommentNode,
				}
			}
			return cell
		}
	}
	return nil
}

func (cells Cells) Value(name string) *html.Node {
	for i := range cells {
		if key := cells.At(i, 0); slugify(key) == name {
			cell := cells.At(i, 1)
			if s := xhtml.TextContent(cell); s == "" {
				// If there's only one column or the column is wide, skip down
				if len(cells[i]) == 1 || xhtml.Attr(key, "colspan") != "" {
					cell = cells.At(i+1, 0)
				}
			}
			if stringx.RemoveAllWhitespace(slugify(cell)) == "n/a" {
				return &html.Node{
					Type: html.CommentNode,
				}
			}
			return cell
		}
	}
	return nil
}

func (cells Cells) Map[T any](f func(*html.Node) T) [][]T {
	rows := make([][]T, 0, len(cells))
	for _, row := range cells {
		rowT := make([]T, 0, len(row))
		for _, col := range row {
			rowT = append(rowT, f(col))
		}
		rows = append(rows, rowT)
	}
	return rows
}
