package blocko_test

import (
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/blocko"
	"github.com/spotlightpa/almanack/internal/testfile"
)

func TestGoldenFiles(t *testing.T) {
	testfile.GlobRun(t, "testdata/*.html", func(path string, t *testing.T) {
		bareName := strings.TrimSuffix(path, ".html")
		in := testfile.Read(t, path)

		got, err := blocko.HTMLToMarkdown(in)
		be.NilErr(t, err)

		testfile.Equal(t, bareName+".md", got)
	})
}
