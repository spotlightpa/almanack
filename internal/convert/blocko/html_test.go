package blocko

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/xhtml"
)

func TestIsEmpty(t *testing.T) {
	type testcase struct {
		in    string
		empty bool
	}
	assert.Run(t, map[string]testcase{
		"span":       {"<span></span>", true},
		"div":        {"<div></div>", false},
		"span-space": {"<span> </span>", true},
		"span-nl":    {"<span>\n\n</span>", true},
		"text-blank": {"<span>\n</span> ", true},
		"text":       {"x", false},
		"span-text":  {"<span></span>x", false},
		"nested":     {"<a><b>\n</b></a> ", true},
		"nested-x":   {"<a><b>x</b></a> ", false},
	}, func(be assert.TB, tc testcase) {
		div := xhtml.New("div")
		be.NilError(xhtml.SetInnerHTML(div, tc.in))
		be.Equal(tc.empty, isEmpty(div))
	})
}
