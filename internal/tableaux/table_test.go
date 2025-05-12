package tableaux_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/tableaux"
	"golang.org/x/net/html"
)

func TestTable(t *testing.T) {
	t.Parallel()
	testfile.Run(t, "testdata/*.html", func(t *testing.T, path string) {
		in := testfile.Read(t, path)
		bareName := testfile.Ext(path, "")

		root, err := html.Parse(strings.NewReader(in))
		be.NilErr(t, err)
		i := 0
		for _, tbl := range tableaux.Tables(root) {
			i++
			rows := tableaux.Map(tbl, xhtml.InnerHTML)
			testfile.EqualJSON(t, fmt.Sprintf("%s-%d.json", bareName, i), &rows)
		}
	})
}
