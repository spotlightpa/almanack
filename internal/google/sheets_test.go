package google_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/pkg/almlog"
	"gopkg.in/Iwark/spreadsheet.v2"
)

func TestSheetToFileObjects(t *testing.T) {
	ctx := context.Background()
	almlog.UseTestLogger(t)

	cl := &http.Client{}
	cl.Transport = requests.Replay("testdata/sheets")

	obj, err := google.SheetToDonorWall(ctx, cl, "abc123")
	be.NilErr(t, err)
	testfile.EqualJSON(t, "testdata/sheets/want.json", obj)
}

func TestSheetMap(t *testing.T) {
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
	be.AllEqual(t, []abc{{"1", "2", "3"}, {"4", "5", "6"}}, got)
}
