package blocko_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/nkotb/pkg/blocko"
)

func read(t testing.TB, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

func TestGoldenFiles(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*.html")
	be.NilErr(t, err)
	for i := range inputs {
		inHTML := inputs[i]
		name := filepath.Base(inHTML)
		name = strings.TrimSuffix(name, ".html")
		t.Run(name, func(t *testing.T) {
			in := strings.NewReader(read(t, inHTML))

			var buf strings.Builder
			blocko.HTMLToMarkdown(&buf, in)

			wantMD := strings.TrimSuffix(inHTML, ".html") + ".md"
			want := read(t, wantMD)
			be.NilErr(t, err)
			got := buf.String()
			be.Debug(t, func() {
				bad := filepath.Join("testdata", name+".xxx.md")
				os.WriteFile(bad, []byte(got), 0644)
			})

			be.Equal(t, want, got)
		})
	}
}
