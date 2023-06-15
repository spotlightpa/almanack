package almanack

import (
	"testing"

	"github.com/carlmjohnson/be"
)

func TestSlugFromURL(t *testing.T) {
	cases := []struct{ in, out string }{
		{},
		{
			in:  "/politics/pennsylvania/spl/republicans-pa-house-control-constitional-amendments-abortion-20221216.html",
			out: "republicans-pa-house-control-constitional-amendments-abortion",
		},
		{
			in:  "weird",
			out: "weird",
		},
	}
	for _, tc := range cases {
		be.Equal(t, tc.out, slugFromURL(tc.in))
	}
}
