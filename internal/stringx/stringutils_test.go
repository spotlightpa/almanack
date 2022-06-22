package stringx_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/stringx"
)

func TestSlugify(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"", ""},
		{"  b  ", "b"},
		{"  ab  ", "ab"},
		{"  a b the c  ", "b-c"},
		{"Pa.'s favorite", "pennsylvanias-favorite"},
		{"the (fort~Nightly)   news  ", "fort-nightly-news"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			be.Equal(t, tc.want, stringx.Slugify(tc.input))
		})
	}
}
