package almsvc

import (
	"path"
	"testing"

	"github.com/earthboundkid/assert"
)

func TestMakeImageName(t *testing.T) {
	cases := map[string]struct {
		ct   string
		want string
	}{
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
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			be := assert.FailNow(t)
			got := makeImageName(tc.ct)
			be.
				Equal(path.Ext(got), tc.want).
				NotMatch(got, `\.\.`)
		})
		var s string
		allocs := testing.AllocsPerRun(10, func() {
			s = makeImageName(tc.ct)
		})
		if allocs > 3 {
			t.Errorf("benchmark regression %q: %v", s, allocs)
		}
	}
}
