package google_test

import (
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/services/google"
	"gopkg.in/Iwark/spreadsheet.v2"
	"net/http"
	"testing"
)

func TestSheetToFileObjects(t *testing.T) {
	be := assert.FailNow(t)
	ctx := t.Context()
	almlog.UseTestLogger(t)

	cl := &http.Client{}
	cl.Transport = reqtest.Replay("testdata/sheets")
	// To update sheet, use this path
	if false {
		svc := google.Service{}
		gcl, err := svc.SheetsClient(ctx)
		be.Zero(err)
		cl.Transport = reqtest.Caching(gcl.Transport, "testdata/sheets")
	}

	obj, err := google.SheetToDonorWall(ctx, cl, "abc123")
	be.Zero(err)
	testfile.EqualJSON(t, "testdata/sheets/want.json", obj)
}

func TestSheetMap(t *testing.T) {
	be := assert.FailNow(t)
	sheet := &spreadsheet.Sheet{
		Rows: [][]spreadsheet.Cell{
			{{Value: "a"}, {Value: "b"}, {Value: "c"}, {}, {}},
			{{Value: "1"}, {Value: "2"}, {Value: "3"}},
			{},
			{{Value: "4"}, {Value: "5"}, {Value: "6"}},
		},
	}
	type abc struct{ A, B, C string }
	got := []abc{}
	sm := google.NewSheetMap(sheet)
	for sm.Next() {
		b := sm.Field("b")
		got = append(got, abc{sm.Field("a"), b, sm.Field("c")})
	}
	be.SlicesEqual(got, []abc{{"1", "2", "3"}, {"4", "5", "6"}})
}
