package almanack

import (
	"testing"

	"github.com/carlmjohnson/be"
)

func TestImageCAS(t *testing.T) {
	cases := []struct {
		body, ct, want string
	}{
		{"", "", "cas/tger-spcf-02s0-9tc0.bin"},
		{"", "image/png", "cas/tger-spcf-02s0-9tc0.png"},
		{"Hello, World!", "image/jpeg", "cas/cpme-4zc8-f4m3-gcdp.jpeg"},
	}
	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := makeCASaddress([]byte(tc.body), tc.ct)
			be.Equal(t, tc.want, got)
		})
		var s string
		body := []byte(tc.body)
		allocs := testing.AllocsPerRun(10, func() {
			s = makeCASaddress(body, tc.ct)
		})
		if allocs > 1 {
			t.Errorf("benchmark regression %q: %v", s, allocs)
		}
	}
}
