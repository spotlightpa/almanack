package blocko

import (
	"bytes"
	"fmt"
	"io"

	"github.com/earthboundkid/xhtml"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	nethtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func Minify(r io.Reader) (*nethtml.Node, error) {
	m := minify.New()
	m.AddFunc("text/html", (&html.Minifier{
		KeepEndTags: true,
	}).Minify)

	var out bytes.Buffer
	if err := m.Minify("text/html", &out, r); err != nil {
		return nil, err
	}
	doc, err := nethtml.Parse(&out)
	if err != nil {
		return nil, err
	}
	body := xhtml.Select(doc, xhtml.WithAtom(atom.Body))
	if body == nil {
		return nil, fmt.Errorf("could not find body")
	}

	return body, nil
}
