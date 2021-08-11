package almanack

import (
	"os"
	"strings"
	"testing"
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
			if !strings.HasSuffix(got, tc.want) {
				t.Errorf("makeImageName(%q) == %q != *%q$", tc.ct, got, tc.want)
			}
			if strings.Contains(got, "..") {
				t.Errorf("makeImageName(%q) == %q", tc.ct, got)
			}
		})
		if os.Getenv("ALMANACK_BENCH_IMAGE_NAMES") == "" {
			continue
		}
		var s string
		r := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				s = makeImageName(tc.ct)
			}
		})
		if r.AllocsPerOp() > 3 {
			t.Errorf("benchmark regression %q: %v", s, r.MemString())
		}
	}
}
