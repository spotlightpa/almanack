package blocko_test

import (
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/testfile"
)

func TestGoldenFiles(t *testing.T) {
	testfile.Run(t, "testdata/*.html", func(t *testing.T, path string) {
		bareName := strings.TrimSuffix(path, ".html")
		in := testfile.Read(t, path)

		got, err := blocko.MinifyAndBlockize(in)
		be.NilErr(t, err)

		testfile.Equal(t, bareName+".md", got)
	})
}
