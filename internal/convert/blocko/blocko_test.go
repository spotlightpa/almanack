package blocko_test

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/convert/blocko"
)

func TestGoldenFiles(t *testing.T) {
	testfile.Run(t, "testdata/*.html", func(t *testing.T, path string) {
		in := testfile.Read(t, path)

		got := assert.FailNow(t).OK(blocko.MinifyAndBlockize(in))

		testfile.Equal(t, testfile.Ext(path, ".md"), got)
	})
}
