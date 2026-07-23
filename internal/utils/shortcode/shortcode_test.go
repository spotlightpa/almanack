package shortcode_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/utils/shortcode"
)

func TestNew(t *testing.T) {
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
		be.Equal(t, tc.want, got)
	}
}

func TestNewPanicsOnOddAttrs(t *testing.T) {
	r := be.Panicked(func() {
		shortcode.New("picture", "src")
	})
	be.Nonzero(t, r)
}
