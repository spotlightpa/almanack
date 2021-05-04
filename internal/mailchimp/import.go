package mailchimp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Import(ctx context.Context, cl *http.Client, page string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, page, nil)
	if err != nil {
		return "", err
	}
	rsp, err := cl.Do(req)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status for %s: %s", page, rsp.Status)
	}
	return PageContent(bufio.NewReader(rsp.Body))
}

func PageContent(r io.Reader) (body string, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		return
	}
	var bNode *html.Node
	if !findNode(doc, func(n *html.Node) bool {
		if n.DataAtom != atom.Body {
			return false
		}
		bNode = n
		return true
	}) {
		err = fmt.Errorf("could not find body element")
		return
	}
	var remove [][2]*html.Node
	visitAll(bNode, func(n *html.Node) {
		if n.Type != html.CommentNode &&
			n.DataAtom != atom.Style &&
			n.DataAtom != atom.Script &&
			!idEq(n, "awesomewrap") {
			return
		}
		if p := n.Parent; p != nil {
			remove = append(remove, [2]*html.Node{p, n})
		}
		return
	})
	for i := range remove {
		p, c := remove[i][0], remove[i][1]
		p.RemoveChild(c)
	}
	var buf strings.Builder
	for n := bNode.FirstChild; n != nil; n = n.NextSibling {
		if err = html.Render(&buf, n); err != nil {
			return
		}
	}
	body = buf.String()
	return
}

func idEq(n *html.Node, id string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, attr := range n.Attr {
		if attr.Key == "id" && attr.Val == id {
			return true
		}
	}
	return false
}

func findNode(n *html.Node, callback func(*html.Node) bool) bool {
	if callback(n) {
		return true
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if findNode(c, callback) {
			return true
		}
	}
	return false
}

func visitAll(n *html.Node, callback func(*html.Node)) {
	findNode(n, func(n *html.Node) bool {
		callback(n)
		return false
	})
}
