package mailchimp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
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
		Handle(requests.ToHTML(&node)).
		Fetch(ctx)
	if err != nil {
		return "", err
	}

	return PageContent(&node)
}

func PageContent(doc *html.Node) (body string, err error) {
	bNode := xhtml.Find(doc, xhtml.WithBody)
	if bNode == nil {
		err = fmt.Errorf("could not find body element")
		return
	}

	remove := xhtml.FindAll(bNode, func(n *html.Node) bool {
		return n.Type == html.CommentNode ||
			n.DataAtom == atom.Style ||
			n.DataAtom == atom.Script ||
			xhtml.Attr(n, "id") == "awesomewrap"
	})

	for _, c := range remove {
		p := c.Parent
		p.RemoveChild(c)
	}

	body = xhtml.ContentsToString(bNode)
	return
}
