package xhtml_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/xhtml"
)

func TestSetInnerHTML(t *testing.T) {
	n := xhtml.New("p")
	be.NilErr(t, xhtml.SetInnerHTML(n, "Hello, <i>World!</i>"))
	be.Equal(t, `<p>Hello, <i>World!</i></p>`, xhtml.ToString(n))

	be.NilErr(t, xhtml.SetInnerHTML(n, "Jello, <i>World!</i>"))
	be.Equal(t, `<p>Jello, <i>World!</i></p>`, xhtml.ToString(n))

	n = xhtml.New("script")
	be.NilErr(t, xhtml.SetInnerHTML(n, "let i = 1 > 2"))
	be.Equal(t, `<script>let i = 1 > 2</script>`, xhtml.ToString(n))

	n = xhtml.New("p")
	be.NilErr(t, xhtml.SetInnerHTML(n, "</a></b>"))
	be.Equal(t, `<p></p>`, xhtml.ToString(n))
}
