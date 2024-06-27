package almanack

import (
	"testing"

	"github.com/carlmjohnson/be"
)

func TestImageCAS(t *testing.T) {
	cases := []struct {
		body, ct, want string
	}{
		{"", "", "cas/tger-spcf-02s0-9tc0.bin"},
		{"", "image/png", "cas/tger-spcf-02s0-9tc0.png"},
		{"Hello, World!", "image/jpeg", "cas/cpme-4zc8-f4m3-gcdp.jpeg"},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := makeCASaddress([]byte(tc.body), tc.ct)
			be.Equal(t, tc.want, got)
		})
		var s string
		body := []byte(tc.body)
		allocs := testing.AllocsPerRun(10, func() {
			s = makeCASaddress(body, tc.ct)
		})
		if allocs > 1 {
			t.Errorf("benchmark regression %q: %v", s, allocs)
		}
	}
}

func TestReplaceSpotlightShortcodes(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", ""},
		{"Hello, World!", "Hello, World!"},
		{
			`<div data-spl-embed-version="2" data-spl-src="https://www.spotlightpa.org/embeds/cta/"></div>`,
			`<div data-spl-embed-version="2" data-spl-src="https://www.spotlightpa.org/embeds/cta/"></div>`,
		},
		{
			`
<script src="https://www.spotlightpa.org/embed.js" async></script>

<div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/cta/"></div>`,
			`{{<embed/cta>}}`,
		},
		{
			`
<script src="https://www.spotlightpa.org/embed.js" async></script>
<div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/cta/"></div>

<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/donate/?teaser_text=a%20a&cta_text=b%20b&eyebrow_text=%22c%22"></div>`,
			`{{<embed/cta>}}
{{<embed/donate cta_text="b b" eyebrow_text="&#34;c&#34;" teaser_text="a a">}}`,
		},
		{
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/donate/?eyebrow_text=BERKS%20BUREAU&cta_text=CONTRIBUTE%20NOW&teaser_text=Make%20a%20gift%20and%20become%20a%20founding%20member%20of%20Spotlight%20PA's%20Berks%20County%20bureau.%20"></div>
		`,
			`{{<embed/donate cta_text="CONTRIBUTE NOW" eyebrow_text="BERKS BUREAU" teaser_text="Make a gift and become a founding member of Spotlight PA&#39;s Berks County bureau. ">}}`,
		},
	}
	for _, tc := range cases {
		t := be.Relaxed(t)
		be.Equal(t, tc.want, replaceSpotlightShortcodes(tc.in))
	}
}
