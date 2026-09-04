package blocko_test

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/convert/blocko"
)

func TestGoldenFiles(t *testing.T) {
	testfile.Run(t, "testdata/*.html", func(t *testing.T, path string) {
		be := assert.FailNow(t)
		in := testfile.Read(t, path)

		got, err := blocko.MinifyAndBlockize(in)
		be.Zero(err)

		testfile.Equal(t, testfile.Ext(path, ".md"), got)
	})
}
