package mailchimp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqhtml"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func ImportPage(ctx context.Context, cl *http.Client, page string) (body string, err error) {
	defer errorx.Trace(&err)

	var node html.Node
	err = requests.
		URL(page).
		Client(cl).
		Handle(reqhtml.To(&node)).
		Fetch(ctx)
	if err != nil {
		return "", err
	}

	return PageContent(&node)
}

func PageContent(doc *html.Node) (body string, err error) {
	// Move <style> tags into <body>. Shove content into @scope rules.
	styleEls := xhtml.SelectSlice(doc, xhtml.WithAtom(atom.Style))
	for _, styleEl := range styleEls {
		styleEl.Parent.RemoveChild(styleEl)
		styleEl.InsertBefore(&html.Node{
			Type: html.TextNode,
			Data: "\n@scope {\n",
		}, styleEl.FirstChild)
		xhtml.AppendText(styleEl, "\n}\n")
	}

	bNode := xhtml.Select(doc, xhtml.WithBody)
	if bNode == nil {
		err = fmt.Errorf("could not find body element")
		return
	}

	remove := xhtml.SelectSlice(bNode, func(n *html.Node) bool {
		return n.Type == html.CommentNode ||
			n.DataAtom == atom.Style ||
			n.DataAtom == atom.Script ||
			xhtml.Attr(n, "id") == "awesomewrap"
	})
	xhtml.RemoveAll(remove)
	for _, styleEl := range styleEls {
		bNode.InsertBefore(styleEl, bNode.FirstChild)
	}
	body = xhtml.InnerHTML(bNode)
	return
}
