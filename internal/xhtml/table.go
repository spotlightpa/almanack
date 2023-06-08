package xhtml

import (
	"strings"

	"github.com/spotlightpa/almanack/internal/stringx"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Tables(root *html.Node, f func(tbl *html.Node, rows TableNodes)) {
	tables := FindAll(root, WithAtom(atom.Table))
	for _, tblNode := range tables {
		var tbl TableNodes
		rows := FindAll(tblNode, WithAtom(atom.Tr))
		for _, row := range rows {
			tds := FindAll(row, func(n *html.Node) bool {
				return n.DataAtom == atom.Td || n.DataAtom == atom.Th
			})
			tbl = append(tbl, tds)
		}
		f(tblNode, tbl)
	}
}

type TableNodes [][]*html.Node

func (rows TableNodes) At(row, col int) *html.Node {
	if row >= len(rows) {
		return &html.Node{Type: html.TextNode}
	}
	r := rows[row]
	if col >= len(r) {
		return &html.Node{Type: html.TextNode}
	}
	return r[col]
}

func slugify(n *html.Node) string {
	return strings.TrimSpace(stringx.RemoveParens(strings.ToLower(InnerText(n))))
}

func (rows TableNodes) Label() string {
	return slugify(rows.At(0, 0))
}

func (rows TableNodes) Value(name string) *html.Node {
	for i := range rows {
		if slugify(rows.At(i, 0)) == name {
			for _, col := range rows[i][1:] {
				if s := InnerText(col); s != "" {
					return col
				}
			}
			return rows.At(i+1, 0)
		}
	}
	return nil
}

func Map[T any](tbl TableNodes, f func(*html.Node) T) [][]T {
	rows := make([][]T, 0, len(tbl))
	for _, row := range tbl {
		rowT := make([]T, 0, len(row))
		for _, col := range row {
			rowT = append(rowT, f(col))
		}
		rows = append(rows, rowT)
	}
	return rows
}
