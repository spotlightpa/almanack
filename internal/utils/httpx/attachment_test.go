package httpx_test

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/utils/httpx"
)

func TestAttachmentName(t *testing.T) {
	type testcase struct {
		in, out string
	}
	assert.RunAll(t, map[string]testcase{
		"blank":   {"", `attachment; filename=""`},
		"plain":   {"a", `attachment; filename=a`},
		"space":   {" a ", `attachment; filename=" a "`},
		"unicode": {"日本語", `attachment; filename*=utf-8''%E6%97%A5%E6%9C%AC%E8%AA%9E`},
	}, func(be assert.TB, tc testcase) {
		be.Equal(httpx.AttachmentName(tc.in), tc.out)
	})
}
