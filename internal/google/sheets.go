package google

import (
	"context"
	"net/http"
	"strings"

	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/pkg/almlog"
	spreadsheet "gopkg.in/Iwark/spreadsheet.v2"
)

func (gsvc *Service) SheetsClient(ctx context.Context) (cl *http.Client, err error) {
	return gsvc.client(ctx, spreadsheet.Scope)
}

type DonorWall struct {
	List []Donor `json:"list"`
}

type Donor struct {
	Display      string `json:"display"`
	Sort         string `json:"sort"`
	URL          string `json:"url"`
	Category     string `json:"category,omitempty"`
	CategoryRank string `json:"categoryRank,omitempty"`
}

// SheetToDonorWall connects to Google Sheets,
// downloads the spreadsheet
// and returns a map from file name to JSON objects
func SheetToDonorWall(ctx context.Context, cl *http.Client, sheetID string) (map[string]DonorWall, error) {
	service := spreadsheet.NewServiceWithClient(cl)
	doc, err := service.FetchSpreadsheet(sheetID)
	if err != nil {
		return nil, resperr.WithUserMessage(err, "Problem fetching Google Sheet")
	}

	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "got sheet", "id", sheetID, "title", doc.Properties.Title)

	sheet, err := doc.SheetByTitle("categories")
	if err != nil {
		return nil, resperr.WithUserMessage(err, "Spreadsheet missing 'categories' sheet.")
	}
	sm := NewSheetMap(sheet)
	cats := make(map[string]string)
	for sm.Next() {
		cname := sm.Field("Name")
		crank := sm.Field("Rank")
		cats[cname] = crank
	}

	sheet, err = doc.SheetByTitle("data")
	if err != nil {
		return nil, resperr.WithUserMessage(err, "Spreadsheet missing 'data' sheet.")
	}

	m := make(map[string]DonorWall)

	sm = NewSheetMap(sheet)
	for sm.Next() {
		sname := sm.Field("Sheet")
		fname := sm.Field("File")
		wallsheet, err := doc.SheetByTitle(sname)
		if err != nil {
			return nil, resperr.WithUserMessagef(err, "Spreadsheet missing %q sheet.", sname)
		}
		var wall DonorWall
		wallmap := NewSheetMap(wallsheet)
		for wallmap.Next() {
			wall.List = append(wall.List, Donor{
				Display:      wallmap.Field("Display"),
				Sort:         wallmap.Field("Sort"),
				URL:          wallmap.Field("URL"),
				Category:     wallmap.Field("Range"),
				CategoryRank: cats[wallmap.Field("Range")],
			})
		}
		m[fname] = wall
	}

	if len(m) < 1 {
		return nil, resperr.WithUserMessage(err, "No rows in 'data' sheet.")
	}
	return m, nil
}

type SheetMap struct {
	sheet *spreadsheet.Sheet
	idx   map[string]int
	row   int
}

func NewSheetMap(sheet *spreadsheet.Sheet) *SheetMap {
	return &SheetMap{sheet, nil, 0}
}

func (sm *SheetMap) Next() bool {
	// Initialize header if empty
	if sm.idx == nil {
		if len(sm.sheet.Rows) < 1 {
			return false
		}
		sm.idx = make(map[string]int)
		for i, cell := range sm.sheet.Rows[0] {
			s := strings.ToLower(strings.TrimSpace(cell.Value))
			if s == "" {
				continue
			}
			sm.idx[s] = i
		}
	}
	// If a row is empty, skip to the next row.
	// Return false once you're out of range.
	for {
		sm.row++
		if len(sm.sheet.Rows) <= sm.row {
			return false
		}
		for _, cell := range sm.sheet.Rows[sm.row] {
			s := strings.TrimSpace(cell.Value)
			if s != "" {
				return true
			}
		}
	}
}

// Field returns the value in the currently loaded row of the column
// corresponding to fieldname.
func (sm *SheetMap) Field(fieldname string) string {
	fieldname = strings.ToLower(fieldname)
	if idx, ok := sm.idx[fieldname]; ok {
		return strings.TrimSpace(sm.sheet.Rows[sm.row][idx].Value)
	}
	return ""
}
