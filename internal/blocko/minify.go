package blocko

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/spotlightpa/almanack/internal/xhtml"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
	nethtml "golang.org/x/net/html"
)

func Minify(s string) (*nethtml.Node, error) {
	m := minify.New()
	m.AddFunc("text/html", (&html.Minifier{
		KeepEndTags: true,
	}).Minify)

	var out bytes.Buffer
	r := strings.NewReader(s)
	if err := m.Minify("text/html", &out, r); err != nil {
		return nil, err
	}
	doc, err := nethtml.Parse(&out)
	if err != nil {
		return nil, err
	}
	body := xhtml.Find(doc, xhtml.WithBody)
	if body == nil {
		return nil, fmt.Errorf("could not find body")
	}

	return body, nil
}
