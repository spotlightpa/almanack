package xhtml_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/xhtml"
)

func TestIsBalanced(t *testing.T) {
	tcs := []struct {
		string
		bool
	}{
		{"", true},
		{"<a></a>", true},
		{"<a><b>hi</b></a>", true},
		{"hello <br /> world", true},
		{"<a><b></b><c></c></a>", true},
		{"</a>", false},
		{"<a></b>", false},
		{"<a><b><c></b></c></a>", false},
	}
	for _, testcase := range tcs {
		t.Run(testcase.string, func(t *testing.T) {
			be.Equal(t, testcase.bool, xhtml.IsBalanced(testcase.string))
		})
	}
}
