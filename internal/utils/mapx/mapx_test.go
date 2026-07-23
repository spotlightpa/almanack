package mapx_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/utils/mapx"
)

func TestFlatten(t *testing.T) {
	cases := []struct {
		in   map[string]string
		want []string
	}{
		{nil, []string{}},
		{
			map[string]string{"a": "1"},
			[]string{"a", "1"},
		},
		{
			// Keys are sorted
			map[string]string{"b": "2", "a": "1"},
			[]string{"a", "1", "b", "2"},
		},
	}
	for _, tc := range cases {
		got := mapx.Flatten(tc.in)
		be.DeepEqual(t, tc.want, got)
	}
}

func TestFlattenMulti(t *testing.T) {
	cases := []struct {
		in   map[string][]string
		want []string
	}{
		{nil, []string{}},
		{
			map[string][]string{"a": {"1"}},
			[]string{"a", "1"},
		},
		{
			// Keys are sorted
			map[string][]string{"b": {"2"}, "a": {"1"}},
			[]string{"a", "1", "b", "2"},
		},
		{
			// Multi-value keys repeat the key
			map[string][]string{"k": {"a", "b"}},
			[]string{"k", "a", "k", "b"},
		},
	}
	for _, tc := range cases {
		got := mapx.FlattenMulti(tc.in)
		be.DeepEqual(t, tc.want, got)
	}
}
