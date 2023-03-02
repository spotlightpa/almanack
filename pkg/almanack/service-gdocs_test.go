package almanack

import (
	"testing"

	"github.com/carlmjohnson/be"
)

func TestGdocsHashPath(t *testing.T) {
	cases := map[string]struct {
		doc, img, ext string
	}{
		"docs/pw1a-jp8n-k7ft-dw3x/pw1a-jp8n-k7ft.":     {"", "", ""},
		"docs/3t9h-5had-dccc-r7wz/7x0z-d5em-s0bk.jpeg": {"a", "b", "jpeg"},
		"docs/3t9h-5had-dccc-r7wz/d9zb-tfes-g9ee.jpeg": {"a", "c", "jpeg"},
		"docs/9m1c-wx3a-5cq1-q1jf/7x0z-d5em-s0bk.png":  {"x", "b", "png"},
		"docs/9m1c-wx3a-5cq1-q1jf/hphp-dwr4-4z40.png":  {"x", "y", "png"},
	}
	for name, tc := range cases {
		got := gdocsHashPath(tc.doc, tc.img, tc.ext)
		be.Equal(be.Relaxed(t), name, got)
	}
}
