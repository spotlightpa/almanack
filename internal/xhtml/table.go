package xhtml

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Tables(root *html.Node, f func(tbl *html.Node, rows TableNodes)) {
	tables := FindAll(root, func(n *html.Node) bool {
		return n.DataAtom == atom.Table
	})
	for _, tblNode := range tables {
		var tbl TableNodes
		rows := FindAll(tblNode, func(n *html.Node) bool {
			return n.DataAtom == atom.Tr
		})
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

func (tbl TableNodes) At(row, col int) *html.Node {
	if row >= len(tbl) {
		return &html.Node{Type: html.TextNode}
	}
	r := tbl[row]
	if col >= len(r) {
		return &html.Node{Type: html.TextNode}
	}
	return r[col]
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
