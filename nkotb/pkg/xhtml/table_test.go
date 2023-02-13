package xhtml_test

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/nkotb/pkg/testfile"
	"github.com/spotlightpa/nkotb/pkg/xhtml"
	"golang.org/x/net/html"
)

func TestTable(t *testing.T) {
	t.Parallel()
	testfile.GlobRun(t, "testdata/*.html", func(path string, t *testing.T) {
		bareName := strings.TrimSuffix(path, ".html")
		want := testfile.Read(t, bareName+".csv")

		in := testfile.Read(t, path)
		root, err := html.Parse(strings.NewReader(in))
		be.NilErr(t, err)
		var buf strings.Builder
		w := csv.NewWriter(&buf)
		xhtml.Tables(root, func(_ *html.Node, tbl xhtml.TableNodes) {
			rows := xhtml.Map(tbl, xhtml.ContentsToString)
			be.NilErr(t, w.WriteAll(rows))
		})
		w.Flush()
		be.NilErr(t, w.Error())
		be.Equal(t, want, buf.String())
	})
}
