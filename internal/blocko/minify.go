package blocko

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spotlightpa/almanack/internal/xhtml"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	nethtml "golang.org/x/net/html"
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
	body := xhtml.Select(doc, xhtml.WithBody)
	if body == nil {
		return nil, fmt.Errorf("could not find body")
	}

	return body, nil
}
