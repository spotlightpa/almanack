package tableaux_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/earthboundkid/xhtml"
	"github.com/spotlightpa/almanack/internal/convert/tableaux"
	"golang.org/x/net/html"
)

func TestTable(t *testing.T) {
	t.Parallel()
	testfile.Run(t, "testdata/*.html", func(t *testing.T, path string) {
		be := assert.FailNow(t)
		in := testfile.Read(t, path)
		bareName := testfile.Ext(path, "")

		root := be.OK(html.Parse(strings.NewReader(in)))
		i := 0
		for _, tbl := range tableaux.Tables(root) {
			i++
			rows := tableaux.Map(tbl, xhtml.InnerHTML)
			testfile.EqualJSON(t, fmt.Sprintf("%s-%d.json", bareName, i), &rows)
		}
	})
}
