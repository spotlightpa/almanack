package blocko

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func prep(r io.Reader) (*xhtml.Node, error) {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)

	var out bytes.Buffer
	if err := m.Minify("text/html", &out, r); err != nil {
		return nil, err
	}
	doc, err := xhtml.Parse(&out)
	if err != nil {
		return nil, err
	}
	body := findNode(doc, func(n *xhtml.Node) *xhtml.Node {
		if n.DataAtom == atom.Body {
			return n
		}
		return nil
	})
	if body == nil {
		return nil, fmt.Errorf("could not find body")
	}
	return body, nil
}
