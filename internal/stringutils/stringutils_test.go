package stringutils_test

import (
	"testing"

	"github.com/spotlightpa/almanack/internal/stringutils"
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
			got := stringutils.Slugify(tc.input)
			if got != tc.want {
				t.Fatalf("got %q", got)
			}
		})
	}
}
