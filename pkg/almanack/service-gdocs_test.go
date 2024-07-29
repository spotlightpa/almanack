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
		{"Hello, World!", `{{<embed/raw srcdoc="Hello, World!">}}`},
		{"<div data-whatever>Hello, World!</div>", `{{<embed/raw srcdoc="&lt;div data-whatever&gt;Hello, World!&lt;/div&gt;">}}`},
		{
			`<div data-spl-embed-version="2" data-spl-src="https://www.spotlightpa.org/embeds/cta/"></div>`,
			`{{<embed/raw srcdoc="&lt;div data-spl-embed-version=&#34;2&#34; data-spl-src=&#34;https://www.spotlightpa.org/embeds/cta/&#34;&gt;&lt;/div&gt;">}}`,
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
		{
			`
			<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/cta/?url=https%3A%2F%2Fcentregives.org%2Forganizations%2F273-spotlight-pa&eyebrow=Centre%20Gives&cta=Contribute%20Now&body=Give%20to%20Spotlight%20PA%20to%20help%20us%20hold%20Penn%20State%20accountable.%20Your%20gift%20goes%20further%20with%20Centre%20Gives%2C%20but%20there's%20only%20one%20day%20left%20to%20donate."></div>
			`,
			`{{<embed/cta body="Give to Spotlight PA to help us hold Penn State accountable. Your gift goes further with Centre Gives, but there&#39;s only one day left to donate." cta="Contribute Now" eyebrow="Centre Gives" url="https://centregives.org/organizations/273-spotlight-pa">}}`,
		},
		{
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/newsletter/?cta=Get%20all%20of%20the%20latest%20stories%20from%20Spotlight%20PA%20and%20top%20headlines%20from%20across%20Pennsylvania%2C%20all%20in%20one%20email%20newsletter."></div>
			`,
			`{{<embed/newsletter cta="Get all of the latest stories from Spotlight PA and top headlines from across Pennsylvania, all in one email newsletter.">}}`,
		},
		{
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/tips/?tip_text=Have%20a%20tip%20about%20Penn%20State%20Health%3F%20We%20want%20to%20hear%20from%20you.%20Write%20to%20us%20via%20the%20form%20below%2C%20or%20mail%20information%20to%20Spotlight%20PA%20State%20College%2C%20210%20W.%20Hamilton%20Ave.%2C%20%23331%2C%20State%20College%2C%20PA%2016801."></div>
			`,
			`{{<embed/tips tip_text="Have a tip about Penn State Health? We want to hear from you. Write to us via the form below, or mail information to Spotlight PA State College, 210 W. Hamilton Ave., #331, State College, PA 16801.">}}`,
		},
		{
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/cta/?body=It's%20%3Cb%3Ecool%3C%2Fb%3E."></div>`,
			`{{<embed/cta body="It&#39;s &lt;b&gt;cool&lt;/b&gt;.">}}`,
		},
		{
			`<div style="padding:56.25% 0 0 0;position:relative;"><iframe src="https://player.vimeo.com/video/990627534?h=89f8de8242&color=ffcb05&title=0&byline=0" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" allow="autoplay; fullscreen; picture-in-picture" allowfullscreen></iframe></div><script src="https://player.vimeo.com/api/player.js"></script>
			`,
			`{{<vimeo id="990627534" secret="89f8de8242" >}}`,
		},
		{
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/cta/?body=It's%20%3Cb%3Ecool%3C%2Fb%3E."></div>
			<div style="padding:56.25% 0 0 0;position:relative;"><iframe src="https://player.vimeo.com/video/990627534?h=89f8de8242&color=ffcb05&title=0&byline=0" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" allow="autoplay; fullscreen; picture-in-picture" allowfullscreen></iframe></div><script src="https://player.vimeo.com/api/player.js"></script>
			`,
			`{{<embed/cta body="It&#39;s &lt;b&gt;cool&lt;/b&gt;.">}}` + "\n" +
				`{{<vimeo id="990627534" secret="89f8de8242" >}}`,
		},
	}
	for _, tc := range cases {
		t := be.Relaxed(t)
		be.Equal(t, tc.want, replaceSpotlightShortcodes(tc.in))
	}
}
