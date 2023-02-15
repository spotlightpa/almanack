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
		wantMD := bareName + ".md"
		want := testfile.Read(be.Relaxed(t), wantMD)

		in := testfile.Read(t, path)
		got, err := blocko.HTMLToMarkdown(in)
		be.NilErr(t, err)

		be.Debug(t, func() {
			badname := bareName + ".xxx.md"
			testfile.Write(t, badname, got)
		})

		be.Equal(t, want, got)
	})
}
