package xhtml_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/testfile"
	"github.com/spotlightpa/almanack/internal/xhtml"
	"golang.org/x/net/html"
)

func TestTable(t *testing.T) {
	t.Parallel()
	testfile.Run(t, "testdata/*.html", func(t *testing.T, path string) {
		in := testfile.Read(t, path)
		bareName := strings.TrimSuffix(path, ".html")

		root, err := html.Parse(strings.NewReader(in))
		be.NilErr(t, err)
		i := 0
		xhtml.Tables(root, func(_ *html.Node, tbl xhtml.TableNodes) {
			i++
			rows := xhtml.Map(tbl, xhtml.ContentsToString)
			testfile.EqualJSON(t, fmt.Sprintf("%s-%d.json", bareName, i), &rows)
		})
	})
}
