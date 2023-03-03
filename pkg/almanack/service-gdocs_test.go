package almanack

import (
	"testing"

	"github.com/carlmjohnson/be"
)

func TestGdocsHashPath(t *testing.T) {
	cases := map[string]struct {
		doc, img, ext string
	}{
		"docs/pw1a-jp8n-k7ft/pw1a-jp8n.":     {"", "", ""},
		"docs/3t9h-5had-dccc/7x0z-d5em.jpeg": {"a", "b", "jpeg"},
		"docs/3t9h-5had-dccc/d9zb-tfes.jpeg": {"a", "c", "jpeg"},
		"docs/9m1c-wx3a-5cq1/7x0z-d5em.png":  {"x", "b", "png"},
		"docs/9m1c-wx3a-5cq1/hphp-dwr4.png":  {"x", "y", "png"},
	}
	for name, tc := range cases {
		got := gdocsHashPath(tc.doc, tc.img, tc.ext)
		be.Equal(be.Relaxed(t), name, got)
	}

}
