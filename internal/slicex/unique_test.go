package slicex_test

import (
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/slicex"
)

func TestUniqueFunc(t *testing.T) {
	for _, tc := range []struct{ have, want string }{
		{"", ""},
		{"1", "1"},
		{"1", "1"},
		{"1 2", "1 2"},
		{"1 2 2 3", "1 2 3"},
		{"1 2 3 2 3 4", "1 2 3 4"},
	} {
		in := strings.Fields(tc.have)
		slicex.UniquesFunc(&in, func(s string) string { return s })
		got := strings.Join(in, " ")
		be.Equal(t, tc.want, got)
	}
}
