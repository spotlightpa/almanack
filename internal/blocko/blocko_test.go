package blocko_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/spotlightpa/almanack/internal/blocko"
)

func TestGoldenFiles(t *testing.T) {
	testfile.Run(t, "testdata/*.html", func(t *testing.T, path string) {
		in := testfile.Read(t, path)

		got, err := blocko.MinifyAndBlockize(in)
		be.NilErr(t, err)

		testfile.Equal(t, testfile.Ext(path, ".md"), got)
	})
}
