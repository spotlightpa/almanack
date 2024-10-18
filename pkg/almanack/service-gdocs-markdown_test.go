package almanack

import (
	"testing"

	"github.com/carlmjohnson/be"
)

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
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/cta/?body=It's%20%3Cb%3Ecool%3C%2Fb%3E."></div>` +
				`<div style="padding:56.25% 0 0 0;position:relative;"><iframe src="https://player.vimeo.com/video/990627534?h=89f8de8242&color=ffcb05&title=0&byline=0" style="position:absolute;top:0;left:0;width:100%;height:100%;" frameborder="0" allow="autoplay; fullscreen; picture-in-picture" allowfullscreen></iframe></div><script src="https://player.vimeo.com/api/player.js"></script>` +
				`<div class="flourish-embed flourish-table" data-src="visualisation/18665405"><script src="https://public.flourish.studio/resources/embed.js"></script></div>` +
				`<iframe title="Most common reports of misconduct at Penn State" aria-label="Bar Chart" id="datawrapper-chart-8hdP8" src="https://datawrapper.dwcdn.net/8hdP8/5/" scrolling="no" frameborder="0" style="width: 0; min-width: 100% !important; border: none;" height="650" data-external="1"></iframe><script type="text/javascript">!function(){"use strict";window.addEventListener("message",(function(a){if(void 0!==a.data["datawrapper-height"]){var e=document.querySelectorAll("iframe");for(var t in a.data["datawrapper-height"])for(var r=0;r<e.length;r++)if(e[r].contentWindow===a.source){var i=a.data["datawrapper-height"][t]+"px";e[r].style.height=i}}}))}();
				</script>` +
				`<iframe class="scribd_iframe_embed" title="GOP Senate letter to Congress" src="https://www.scribd.com/embeds/489945100/content?start_page=1&view_mode=scroll&access_key=key-29gXRO2IEpbetVJqdcQL" data-auto-height="true" data-aspect-ratio="0.7729220222793488" scrolling="no" width="100%" height="600" frameborder="0"></iframe><p  style="   margin: 12px auto 6px auto;   font-family: Helvetica,Arial,Sans-serif;   font-style: normal;   font-variant: normal;   font-weight: normal;   font-size: 14px;   line-height: normal;   font-size-adjust: none;   font-stretch: normal;   -x-system-font: none;   display: block;"   ><a title="View GOP Senate letter to Congress on Scribd" href="https://www.scribd.com/document/489945100/GOP-Senate-letter-to-Congress"  style="text-decoration: underline;">GOP Senate letter to Congress</a> by <a title="View Sarah Anne Hughes's profile on Scribd" href="https://www.scribd.com/user/507961525/Sarah-Anne-Hughes"  style="text-decoration: underline;">Sarah Anne Hughes</a></p>` +
				`<div data-tf-live="01HFS5TPTDZNNK7PV48PN752KC"></div><script src="//embed.typeform.com/next/embed.js"></script>`,
			`{{<embed/cta body="It&#39;s &lt;b&gt;cool&lt;/b&gt;.">}}` + "\n" +
				`{{<vimeo id="990627534" secret="89f8de8242" >}}` + "\n" +
				`{{<flourish src="visualisation/18665405" >}}` + "\n" +
				`{{<datawrapper src="https://datawrapper.dwcdn.net/8hdP8/5/" height="650" >}}` + "\n" +
				`{{<scribd src="https://www.scribd.com/embeds/489945100/content?start_page=1&amp;view_mode=scroll&amp;access_key=key-29gXRO2IEpbetVJqdcQL" >}}` + "\n" +
				`{{<typeform id="01HFS5TPTDZNNK7PV48PN752KC" >}}`,
		},
		{
			`
		<a href="#XHEPNKWD" style="display: none"></a>
		`,
			`{{<fundraiseup id="XHEPNKWD">}}`,
		},
		{
			`<a href="#XMXVXGPU" style="display: none"></a>`,
			`{{<fundraiseup id="XMXVXGPU">}}`,
		},
		{
			`<iframe width="100%" height="315" src="https://www.youtube.com/embed/XbnubJm-ofk?si=Umnc0yFxgFLJ9uyK" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>`,
			`{{<youtube id="XbnubJm-ofk" loading="lazy">}}`,
		},
		{
			`<iframe width="560" height="315" src="https://www.youtube-nocookie.com/embed/XbnubJm-ofk?si=QdOwK7Cv4oF3QNb_&amp;start=11" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" referrerpolicy="strict-origin-when-cross-origin" allowfullscreen></iframe>`,
			`{{<youtube id="XbnubJm-ofk" start="11" loading="lazy">}}`,
		},
		{
			`<script type="text/javascript">
			(function (e, o) {
			  var deweyConfig = {
				key: "3a59cd0d-c005-4a65-807d-3a5e539bd515",
				options: {
				  targetElementId: "deweyChatTarget",
				}
			  };
			  var n = window.dewey = window.dewey || {}; if (n.invoked) { console.error("Dewey snippet included twice."); return } n.invoked = true; n.load = function (e, t) { return new Promise(((r, d) => { var i = o.createElement("script"); i.type = "text/javascript"; i.async = true; i.onload = r; i.onerror = d; i.src = ` + "`" + `https://app.askdewey.co/dewey.js/v1/${e}/dewey.min.js` + "`" + `; n._loadOptions = t; o.head.appendChild(i) })) }; n.SNIPPET_VERSION = "0.0.2"; async function t() { try { await n.load(deweyConfig.key, deweyConfig.options); n.start() } catch (e) { console.error("Failed to load Dewey script:", e) } } t()
			})(window, document);
		  </script>
		  <div id="deweyChatTarget"></div>`,
			`{{<dewey-assistant>}}`,
		},
	}
	for _, tc := range cases {
		t := be.Relaxed(t)
		be.Equal(t, tc.want, replaceSpotlightShortcodes(tc.in))
	}
}
