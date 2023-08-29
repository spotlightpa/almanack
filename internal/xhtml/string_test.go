package xhtml_test

import (
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestInnerText(t *testing.T) {
	s := `<p><em></em>&lt;div class=&#34;flourish-embed flourish-cards&#34; data-src=&#34;visualisation/14836391&#34;&gt;&lt;script src=&#34;<a href="https://public.flourish.studio/resources/embed.js">https://public.flourish.studio/resources/embed.js</a>&#34;&gt;&lt;/script&gt;&lt;/div&gt;
	</p>`
	doc, err := html.Parse(strings.NewReader(s))
	be.NilErr(t, err)
	p := xhtml.Find(doc, xhtml.WithAtom(atom.P))
	got := xhtml.InnerText(p)
	be.Equal(t, `<div class="flourish-embed flourish-cards" data-src="visualisation/14836391"><script src="https://public.flourish.studio/resources/embed.js"></script></div>`, got)
}
