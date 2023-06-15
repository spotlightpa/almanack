package slicex_test

import (
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/slicex"
	"golang.org/x/exp/slices"
)

func TestDeleteFunc(t *testing.T) {
	for _, tc := range []struct{ in, out []string }{
		{},
		{in: []string{"A", "B", "C"}, out: []string{"A", "B", "C"}},
		{in: []string{"A", "B", "c"}, out: []string{"A", "B"}},
		{in: []string{"A", "b", "C"}, out: []string{"A", "C"}},
		{in: []string{"A", "b", "c"}, out: []string{"A"}},
		{in: []string{"a", "B", "C"}, out: []string{"B", "C"}},
		{in: []string{"a", "B", "c"}, out: []string{"B"}},
		{in: []string{"a", "b", "C"}, out: []string{"C"}},
		{in: []string{"a", "b", "c"}, out: []string{}},
	} {
		t.Run(strings.Join(tc.in, ""), func(t *testing.T) {
			got := slices.Clone(tc.in)
			slicex.DeleteFunc(&got, func(s string) bool {
				return strings.ToLower(s) == s
			})
			be.AllEqual(t, tc.out, got)
		})
	}
}
