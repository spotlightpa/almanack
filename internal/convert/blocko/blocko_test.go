package blocko_test

import (
	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/convert/blocko"
	"testing"
)

func TestGoldenFiles(t *testing.T) {
	be := assert.FailNow(t)
	testfile.Run(t, "testdata/*.html", func(t *testing.T, path string) {
		in := testfile.Read(t, path)

		got, err := blocko.MinifyAndBlockize(in)
		be.Zero(err)

		testfile.Equal(t, testfile.Ext(path, ".md"), got)
	})
}
