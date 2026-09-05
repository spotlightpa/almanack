package almsvc

import (
	"path"
	"testing"

	"github.com/earthboundkid/assert"
)

func TestMakeImageName(t *testing.T) {
	type testcase struct {
		ct   string
		want string
	}
	cases := map[string]testcase{
		"none":      {"", ".bin"},
		"slash":     {"/", ".bin"},
		"no slash":  {"hello", ".bin"},
		"malformed": {"image/", ".bin"},
		"png":       {"image/png", ".png"},
		"jpeg":      {"image/jpeg", ".jpeg"},
		"tiff":      {"image/tiff", ".tiff"},
		"json":      {"application/json", ".json"},
		"text":      {"text/plain", ".plain"},
	}
	assert.Run(t, cases, func(be assert.TB, tc testcase) {
		got := makeImageName(tc.ct)
		be.
			Equal(path.Ext(got), tc.want).
			NotMatch(got, `\.\.`)
	})
	for _, tc := range cases {
		var s string
		allocs := testing.AllocsPerRun(10, func() {
			s = makeImageName(tc.ct)
		})
		if allocs > 3 {
			t.Errorf("benchmark regression %q: %v", s, allocs)
		}
	}
}
