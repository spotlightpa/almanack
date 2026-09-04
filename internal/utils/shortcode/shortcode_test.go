package shortcode_test

import (
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/utils/shortcode"
	"testing"
)

func TestNew(t *testing.T) {
	be := assert.FailNow(t)
	cases := []struct {
		tag   string
		attrs []string
		want  string
	}{
		{
			tag:  "picture",
			want: `{{<picture>}}`,
		},
		{
			tag:   "picture",
			attrs: []string{"src", "foo.jpg"},
			want:  `{{<picture src="foo.jpg">}}`,
		},
		{
			tag:   "picture",
			attrs: []string{"src", "foo.jpg", "alt", "a dog"},
			want:  `{{<picture src="foo.jpg" alt="a dog">}}`,
		},
		{
			// HTML special chars in values are escaped
			tag:   "embed/raw",
			attrs: []string{"srcdoc", `<b>bold</b>`},
			want:  `{{<embed/raw srcdoc="&lt;b&gt;bold&lt;/b&gt;">}}`,
		},
		{
			// Newlines in values are escaped
			tag:   "embed/raw",
			attrs: []string{"srcdoc", "line1\nline2"},
			want:  `{{<embed/raw srcdoc="line1&#10;line2">}}`,
		},
		{
			// Double quotes in values are escaped
			tag:   "picture",
			attrs: []string{"caption", `say "hi"`},
			want:  `{{<picture caption="say &#34;hi&#34;">}}`,
		},
	}
	for _, tc := range cases {
		got := shortcode.New(tc.tag, tc.attrs...)
		be.Equal(got, tc.want)
	}
}

func TestNewPanicsOnOddAttrs(t *testing.T) {
	be := assert.FailNow(t)
	r := assert.Catch(func() {
		shortcode.New("picture", "src")
	})
	be.NotZero(r)
}
