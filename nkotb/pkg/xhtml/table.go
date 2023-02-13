package xhtml

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Tables(root *html.Node, f func(tbl *html.Node, rows Table)) {
	tables := FindAll(root, func(n *html.Node) bool {
		return n.DataAtom == atom.Table
	})
	for _, tbl := range tables {
		var text Table
		rows := FindAll(tbl, func(n *html.Node) bool {
			return n.DataAtom == atom.Tr
		})
		for _, row := range rows {
			var rowText []string
			tds := FindAll(row, func(n *html.Node) bool {
				return n.DataAtom == atom.Td || n.DataAtom == atom.Th
			})
			for _, td := range tds {
				rowText = append(rowText, ContentsToString(td))
			}
			text = append(text, rowText)
		}
		f(tbl, text)
	}
}

type Table [][]string

func (tbl Table) At(row, col int) string {
	if row >= len(tbl) {
		return ""
	}
	r := tbl[row]
	if col >= len(r) {
		return ""
	}
	return r[col]
}
