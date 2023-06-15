package almanack

import (
	"path"
	"testing"

	"github.com/carlmjohnson/be"
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
			got := makeImageName(tc.ct)
			be.Equal(t, tc.want, path.Ext(got))
			be.NotIn(t, "..", got)
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
