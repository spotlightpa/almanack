package almsvc

import (
	"testing"

	"github.com/earthboundkid/assert"
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
			`{{<vimeo id="990627534" secret="89f8de8242">}}`,
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
				`{{<vimeo id="990627534" secret="89f8de8242">}}` + "\n" +
				`{{<flourish src="visualisation/18665405">}}` + "\n" +
				`{{<datawrapper src="https://datawrapper.dwcdn.net/8hdP8/5/" height="650">}}` + "\n" +
				`{{<scribd src="https://www.scribd.com/embeds/489945100/content?start_page=1&amp;view_mode=scroll&amp;access_key=key-29gXRO2IEpbetVJqdcQL">}}` + "\n" +
				`{{<typeform id="01HFS5TPTDZNNK7PV48PN752KC">}}`,
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
			`{{<youtube id="XbnubJm-ofk" loading="lazy" start="11">}}`,
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
		{
			`<script type="text/javascript">
var deweyConfig = {
  key: "ba7fd845-222f-4d30-b2d1-ef3210971999",
  options: {
    verbose: true,
    targetElementId: "dewey--frame"
  }
};
!function(e,o,n){async function d(){try{await t.load(n.key,n.options),t.start()}catch(e){console.error("Failed to load Dewey script:",e)}}var t=window.dewey=window.dewey||{};t.invoked?console.error("Dewey snippet included twice."):(t.invoked=!0,t.load=function(e,n){return new Promise((e,d)=>{var a=o.createElement("script");a.type="text/javascript",a.async=!0,a.onload=e,a.onerror=d,a.src="https://app.askdewey.co/dewey.js/v1/ba7fd845-222f-4d30-b2d1-ef3210971999/dewey.min.js",t._loadOptions=n,o.head.appendChild(a)})},t.SNIPPET_VERSION="0.0.3",d())}(window,document,deweyConfig);
</script>
<div id="dewey--frame"></div>
`,
			`{{<dewey-assistant>}}`,
		},
		{
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/donate/"></div>`,
			"{{<embed/donate>}}",
		},
		{
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/tips/?tip_text=Do%20you%20have%20information%20relating%20to%20this%20case%20that%20we%20should%20investigate%3F%20Tell%20us%20now.&form_name=statecollege-embed"></div>
		`,
			`{{<embed/tips form_name="statecollege-embed" tip_text="Do you have information relating to this case that we should investigate? Tell us now.">}}`,
		},
		{
			`<script src="https://www.spotlightpa.org/embed.js" async></script><div data-spl-embed-version="1" data-spl-src="https://www.spotlightpa.org/embeds/newsletter/?cta=Sign%20up%20for%20our%20new%20regional%20newsletter%2C%20%3Cb%3ETalk%20of%20the%20Town%3C%2Fb%3E%2C%20and%20get%20all%20the%20news%20and%20notes%20from%20State%20College%20and%20north-central%20PA.&button=Sign%20Up%20Now&preselect=state_college&eyebrow=DON'T%20MISS%20A%20BEAT"></div>
		`,
			`{{<embed/newsletter button="Sign Up Now" cta="Sign up for our new regional newsletter, &lt;b&gt;Talk of the Town&lt;/b&gt;, and get all the news and notes from State College and north-central PA." eyebrow="DON&#39;T MISS A BEAT" preselect="state_college">}}`,
		},
		{
			`<div style="min-height:609px;" id="datawrapper-vis-R2dcp"><script type="text/javascript" defer src="https://datawrapper.dwcdn.net/R2dcp/embed.js" charset="utf-8" data-target="#datawrapper-vis-R2dcp"></script><noscript><img src="https://datawrapper.dwcdn.net/R2dcp/full.png" alt="" /></noscript></div>`,
			`{{<datawrapper src="https://datawrapper.dwcdn.net/R2dcp/" height="609">}}`,
		},
		{
			`<blockquote class="text-post-media" data-text-post-permalink="https://www.threads.com/t/DZ1PHiUiYlx?utm_source=th_embed&amp;utm_campaign=threads_api" data-text-post-version="0" id="ig-tp-DZ1PHiUiYlx" style=" background:#FFF; border-width: 1px; border-style: solid; border-color: #00000026; border-radius: 16px; max-width:658px; margin: 1px; min-width:270px; padding:0; width:99.375%; width:-webkit-calc(100% - 2px); width:calc(100% - 2px);" data-theme="light"> <a href="https://www.threads.com/t/DZ1PHiUiYlx?utm_source=th_embed&amp;utm_campaign=threads_api" style=" background:#FFFFFF; line-height:0; padding:0 0; text-align:center; text-decoration:none; width:100%; font-family: -apple-system, BlinkMacSystemFont, sans-serif;" target="_blank"> <div style=" padding: 40px; display: flex; flex-direction: column; align-items: center;">
<div style=" display:block; height:32px; width:32px; padding-bottom:16px;"> <svg aria-label="Threads" height="32px" role="img" viewbox="0 0 192 192" width="32px" xmlns="http://www.w3.org/2000/svg"> <path d="M141.537 88.9883C140.71 88.5919 139.87 88.2104 139.019 87.8451C137.537 60.5382 122.616 44.905 97.5619 44.745C97.4484 44.7443 97.3355 44.7443 97.222 44.7443C82.2364 44.7443 69.7731 51.1409 62.102 62.7807L75.881 72.2328C81.6116 63.5383 90.6052 61.6848 97.2286 61.6848C97.3051 61.6848 97.3819 61.6848 97.4576 61.6855C105.707 61.7381 111.932 64.1366 115.961 68.814C118.893 72.2193 120.854 76.925 121.825 82.8638C114.511 81.6207 106.601 81.2385 98.145 81.7233C74.3247 83.0954 59.0111 96.9879 60.0396 116.292C60.5615 126.084 65.4397 134.508 73.775 140.011C80.8224 144.663 89.899 146.938 99.3323 146.423C111.79 145.74 121.563 140.987 128.381 132.296C133.559 125.696 136.834 117.143 138.28 106.366C144.217 109.949 148.617 114.664 151.047 120.332C155.179 129.967 155.42 145.8 142.501 158.708C131.182 170.016 117.576 174.908 97.0135 175.059C74.2042 174.89 56.9538 167.575 45.7381 153.317C35.2355 139.966 29.8077 120.682 29.6052 96C29.8077 71.3178 35.2355 52.0336 45.7381 38.6827C56.9538 24.4249 74.2039 17.11 97.0132 16.9405C119.988 17.1113 137.539 24.4614 149.184 38.788C154.894 45.8136 159.199 54.6488 162.037 64.9503L178.184 60.6422C174.744 47.9622 169.331 37.0357 161.965 27.974C147.036 9.60668 125.202 0.195148 97.0695 0H96.9569C68.8816 0.19447 47.2921 9.6418 32.7883 28.0793C19.8819 44.4864 13.2244 67.3157 13.0007 95.9325L13 96L13.0007 96.0675C13.2244 124.684 19.8819 147.514 32.7883 163.921C47.2921 182.358 68.8816 191.806 96.9569 192H97.0695C122.03 191.827 139.624 185.292 154.118 170.811C173.081 151.866 172.51 128.119 166.26 113.541C161.776 103.087 153.227 94.5962 141.537 88.9883ZM98.4405 129.507C88.0005 130.095 77.1544 125.409 76.6196 115.372C76.2232 107.93 81.9158 99.626 99.0812 98.6368C101.047 98.5234 102.976 98.468 104.871 98.468C111.106 98.468 116.939 99.0737 122.242 100.233C120.264 124.935 108.662 128.946 98.4405 129.507Z"></path></svg>
</div>
<div style=" font-size: 15px; line-height: 21px; color: #000000; font-weight: 600; "> View on Threads</div>
</div></a>
</blockquote>
<script async src="https://www.threads.com/embed.js"></script>`,
			`{{<threads id="DZ1PHiUiYlx">}}`,
		},
		{
			`<blockquote class="text-post-media" data-text-post-permalink="https://www.threads.com/t/DZ1PHiUiYlx?utm_source=th_embed&amp;utm_campaign=threads_api" data-text-post-version="0" id="ig-tp-DZ1PHiUiYlx" style=" background:#FFF; border-width: 1px; border-style: solid; border-color: #00000026; border-radius: 16px; max-width:658px; margin: 1px; min-width:270px; padding:0; width:99.375%; width:-webkit-calc(100% - 2px); width:calc(100% - 2px);" data-theme="light"> <a href="https://www.threads.com/t/DZ1PHiUiYlx?utm_source=th_embed&amp;utm_campaign=threads_api" style=" background:#FFFFFF; line-height:0; padding:0 0; text-align:center; text-decoration:none; width:100%; font-family: -apple-system, BlinkMacSystemFont, sans-serif;" target="_blank"> <div style=" padding: 40px; display: flex; flex-direction: column; align-items: center;">
<div style=" display:block; height:32px; width:32px; padding-bottom:16px;"> <svg aria-label="Threads" height="32px" role="img" viewbox="0 0 192 192" width="32px" xmlns="http://www.w3.org/2000/svg"> <path d="M141.537 88.9883C140.71 88.5919 139.87 88.2104 139.019 87.8451C137.537 60.5382 122.616 44.905 97.5619 44.745C97.4484 44.7443 97.3355 44.7443 97.222 44.7443C82.2364 44.7443 69.7731 51.1409 62.102 62.7807L75.881 72.2328C81.6116 63.5383 90.6052 61.6848 97.2286 61.6848C97.3051 61.6848 97.3819 61.6848 97.4576 61.6855C105.707 61.7381 111.932 64.1366 115.961 68.814C118.893 72.2193 120.854 76.925 121.825 82.8638C114.511 81.6207 106.601 81.2385 98.145 81.7233C74.3247 83.0954 59.0111 96.9879 60.0396 116.292C60.5615 126.084 65.4397 134.508 73.775 140.011C80.8224 144.663 89.899 146.938 99.3323 146.423C111.79 145.74 121.563 140.987 128.381 132.296C133.559 125.696 136.834 117.143 138.28 106.366C144.217 109.949 148.617 114.664 151.047 120.332C155.179 129.967 155.42 145.8 142.501 158.708C131.182 170.016 117.576 174.908 97.0135 175.059C74.2042 174.89 56.9538 167.575 45.7381 153.317C35.2355 139.966 29.8077 120.682 29.6052 96C29.8077 71.3178 35.2355 52.0336 45.7381`,
			`{{<threads id="DZ1PHiUiYlx">}}`,
		},
	}
	for _, tc := range cases {
		assert.Continues(t).Equal(replaceSpotlightShortcodes(tc.in), tc.want)
	}
}
