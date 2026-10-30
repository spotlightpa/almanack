package layouts

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/earthboundkid/xhtml"
	"golang.org/x/net/html"
)

// ToHTML executes t with data and parses the resulting HTML into a container div *html.Node.
func ToHTML(t *template.Template, data any) (*html.Node, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("layouts.Render: template %q: %w", t.Name(), err)
	}
	root := xhtml.New("div")
	if err := xhtml.SetInnerHTML(root, buf.String()); err != nil {
		return nil, fmt.Errorf("layouts.Render: template %q produced invalid HTML: %w", t.Name(), err)
	}
	return root, nil
}
