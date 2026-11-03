package blocko_test

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/convert/blocko"
)

func TestGoldenFiles(t *testing.T) {
	testfile.Run(t, "testdata/*.html", func(be assert.TB, path string) {
		in := testfile.Read(be, path)

		got := be.OK(blocko.MinifyAndBlockize(in))

		testfile.Equal(be, testfile.Ext(path, ".md"), got)
	})
}
