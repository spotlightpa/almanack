package xhtml_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html/atom"
)

func TestSetInnerHTML(t *testing.T) {
	n := xhtml.New("p")
	be.NilErr(t, xhtml.SetInnerHTML(n, "Hello, <i>World!</i>"))
	be.Equal(t, `<p>Hello, <i>World!</i></p>`, xhtml.OuterHTML(n))

	be.NilErr(t, xhtml.SetInnerHTML(n, "Jello, <i>World!</i>"))
	be.Equal(t, `<p>Jello, <i>World!</i></p>`, xhtml.OuterHTML(n))

	n = xhtml.New("script")
	be.NilErr(t, xhtml.SetInnerHTML(n, "let i = 1 > 2"))
	be.Equal(t, `<script>let i = 1 > 2</script>`, xhtml.OuterHTML(n))

	n = xhtml.New("p")
	be.NilErr(t, xhtml.SetInnerHTML(n, "</a></b>"))
	be.Equal(t, `<p></p>`, xhtml.OuterHTML(n))
}

func TestUnnestChildren(t *testing.T) {
	n := xhtml.New("div")
	be.NilErr(t, xhtml.SetInnerHTML(n,
		`<a><b><i>test</i> <i>one</i> <em><i>two</i></em> </b></a>`))

	{
		clone := xhtml.Clone(n)
		i := xhtml.Find(clone, xhtml.WithAtom(atom.I))
		xhtml.UnnestChildren(i)
		be.Equal(t, `<a><b>test <i>one</i> <em><i>two</i></em> </b></a>`,
			xhtml.InnerHTML(clone))
	}
	{
		clone := xhtml.Clone(n)
		em := xhtml.Find(clone, xhtml.WithAtom(atom.Em))
		xhtml.UnnestChildren(em)
		be.Equal(t, `<a><b><i>test</i> <i>one</i> <i>two</i> </b></a>`,
			xhtml.InnerHTML(clone))
	}
	{
		clone := xhtml.Clone(n)
		a := xhtml.Find(clone, xhtml.WithAtom(atom.A))
		xhtml.UnnestChildren(a)
		be.Equal(t, `<b><i>test</i> <i>one</i> <em><i>two</i></em> </b>`,
			xhtml.InnerHTML(clone))
	}
	{
		clone := xhtml.Clone(n)
		b := xhtml.Find(clone, xhtml.WithAtom(atom.B))
		xhtml.UnnestChildren(b)
		be.Equal(t, `<a><i>test</i> <i>one</i> <em><i>two</i></em> </a>`,
			xhtml.InnerHTML(clone))
	}
	{
		clone := xhtml.Clone(n)
		is := xhtml.FindAll(clone, xhtml.WithAtom(atom.I))
		for _, n := range is {
			xhtml.UnnestChildren(n)
		}
		be.Equal(t, `<a><b>test one <em>two</em> </b></a>`,
			xhtml.InnerHTML(clone))
	}
}
