package google

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/earthboundkid/resperr/v2"
	spreadsheet "gopkg.in/Iwark/spreadsheet.v2"
)

type MapPage struct {
	Slug        string
	Section     string
	Headline    string
	Eyebrow     string
	Dek         string
	Byline      string
	Date        string
	PublishedAt time.Time
	GeoJSON     string
	Color       string
	Blurb       string
	Description string
	InternalID  string
	Kicker      string
	Topics      []string
	Body        string
	Credits     []MapCredit
}

type MapCredit struct {
	Eyebrow string
	Name    string
	Role    string
	Email   string
}

func (m MapPage) FilePath() string {
	section := m.Section
	if section == "" {
		section = "news"
	}
	return fmt.Sprintf("content/%s/%s.md", section, m.Slug)
}

func (m MapPage) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString("+++\n")
	if m.Byline != "" {
		parts := splitByline(m.Byline)
		sb.WriteString("authors = [")
		for i, p := range parts {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%q", p)
		}
		sb.WriteString("]\n")
	}
	if m.Blurb != "" {
		fmt.Fprintf(&sb, "blurb = %q\n", m.Blurb)
	}
	if m.Byline != "" {
		fmt.Fprintf(&sb, "byline = %q\n", m.Byline)
	}
	if m.Description != "" {
		fmt.Fprintf(&sb, "description = %q\n", m.Description)
	}
	if m.InternalID != "" {
		fmt.Fprintf(&sb, "internal-id = %q\n", m.InternalID)
	}
	if m.Kicker != "" {
		fmt.Fprintf(&sb, "kicker = %q\n", m.Kicker)
	}
	sb.WriteString("layout = \"searchable-map\"\n")
	if !m.PublishedAt.IsZero() {
		fmt.Fprintf(&sb, "published = %s\n", m.PublishedAt.Format(time.RFC3339))
	}
	if m.Slug != "" {
		fmt.Fprintf(&sb, "slug = %q\n", m.Slug)
	}
	sb.WriteString("suppress-ads = true\n")
	if m.Headline != "" {
		fmt.Fprintf(&sb, "title = %q\n", m.Headline)
		fmt.Fprintf(&sb, "title-tag = %q\n", m.Headline)
	}
	if len(m.Topics) > 0 {
		sb.WriteString("topics = [")
		for i, t := range m.Topics {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "%q", t)
		}
		sb.WriteString("]\n")
	}
	sb.WriteString("+++\n\n")

	sb.WriteString("{{<featured/map-header\n")
	if m.Eyebrow != "" {
		fmt.Fprintf(&sb, "  eyebrow=%q\n", m.Eyebrow)
	}
	if m.Headline != "" {
		fmt.Fprintf(&sb, "  hed=%q\n", m.Headline)
	}
	if m.Dek != "" {
		fmt.Fprintf(&sb, "  dek=%q\n", m.Dek)
	}
	if m.Date != "" {
		fmt.Fprintf(&sb, "  date=%q\n", m.Date)
	}
	if m.Byline != "" {
		fmt.Fprintf(&sb, "  byline=%q\n", m.Byline)
	}
	if m.Color != "" {
		fmt.Fprintf(&sb, "  color=%q\n", m.Color)
	}
	sb.WriteString("  outlet=\"Spotlight PA\"\n")
	if m.GeoJSON != "" {
		fmt.Fprintf(&sb, "  geojson=%q\n", m.GeoJSON)
	}
	sb.WriteString(">}}\n")
	if m.Body != "" {
		sb.WriteString(m.Body)
		sb.WriteString("\n")
	}
	sb.WriteString("{{</featured/map-header>}}\n\n")

	if len(m.Credits) > 0 {
		sb.WriteString("{{<featured/footer>}}\n")
		for _, c := range m.Credits {
			sb.WriteString("{{<featured/credit\n")
			if c.Eyebrow != "" {
				fmt.Fprintf(&sb, "  eyebrow=%q\n", c.Eyebrow)
			}
			if c.Name != "" {
				fmt.Fprintf(&sb, "  name=%q\n", c.Name)
			}
			if c.Role != "" {
				fmt.Fprintf(&sb, "  role=%q\n", c.Role)
			}
			if c.Email != "" {
				fmt.Fprintf(&sb, "  email=%q\n", c.Email)
			}
			sb.WriteString(">}}\n")
		}
		sb.WriteString("{{</featured/footer>}}\n")
	}

	return sb.String()
}

func splitByline(byline string) []string {
	byline = strings.TrimPrefix(byline, "By ")
	byline = strings.TrimPrefix(byline, "by ")
	parts := strings.Split(byline, " and ")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

type sheetMapSkipDescription struct {
	sheet *spreadsheet.Sheet
	idx   map[string]int
	row   int
}

func newSheetMapSkipDescription(sheet *spreadsheet.Sheet) *sheetMapSkipDescription {
	return &sheetMapSkipDescription{sheet: sheet, row: 0}
}

func (sm *sheetMapSkipDescription) Next() bool {
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
	for {
		sm.row++
		if len(sm.sheet.Rows) <= sm.row {
			return false
		}
		// skip the description row (row index 1)
		if sm.row == 1 {
			continue
		}
		for _, cell := range sm.sheet.Rows[sm.row] {
			if strings.TrimSpace(cell.Value) != "" {
				return true
			}
		}
	}
}

func (sm *sheetMapSkipDescription) Field(fieldname string) string {
	fieldname = strings.ToLower(fieldname)
	if idx, ok := sm.idx[fieldname]; ok {
		if idx < len(sm.sheet.Rows[sm.row]) {
			return strings.TrimSpace(sm.sheet.Rows[sm.row][idx].Value)
		}
	}
	return ""
}

func SheetToMapPages(ctx context.Context, cl *http.Client, sheetID string) ([]MapPage, error) {
	service := spreadsheet.NewServiceWithClient(cl)
	doc, err := service.FetchSpreadsheet(sheetID)
	if err != nil {
		return nil, resperr.E{E: err, M: "Problem fetching map config sheet"}
	}

	headerSheet, err := doc.SheetByTitle("Header")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Header' sheet"}
	}

	settingsSheet, err := doc.SheetByTitle("Map Settings")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Map Settings' sheet"}
	}

	dataSheet, err := doc.SheetByTitle("Map Data")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Map Data' sheet"}
	}

	creditsSheet, err := doc.SheetByTitle("Credits")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Credits' sheet"}
	}

	hdr := newSheetMapSkipDescription(headerSheet)
	set := newSheetMapSkipDescription(settingsSheet)
	dat := newSheetMapSkipDescription(dataSheet)

	if !hdr.Next() {
		return nil, resperr.E{M: "No data rows in Header sheet"}
	}
	if !set.Next() {
		return nil, resperr.E{M: "No data rows in Map Settings sheet"}
	}
	if !dat.Next() {
		return nil, resperr.E{M: "No data rows in Map Data sheet"}
	}

	slug := hdr.Field("Slug")
	if slug == "" {
		return nil, resperr.E{M: "Header sheet missing Slug value"}
	}

	publishedStr := hdr.Field("Published")
	var publishedAt time.Time
	if publishedStr != "" {
		publishedAt, _ = time.Parse("2006-01-02", publishedStr)
	}
	if publishedAt.IsZero() {
		publishedAt = time.Now()
	}

	topicsStr := hdr.Field("Topics")
	var topics []string
	for _, t := range strings.Split(topicsStr, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			topics = append(topics, t)
		}
	}

	var credits []MapCredit
	cred := newSheetMapSkipDescription(creditsSheet)
	for cred.Next() {
		name := cred.Field("Name")
		if name == "" {
			continue
		}
		credits = append(credits, MapCredit{
			Eyebrow: cred.Field("Eyebrow"),
			Name:    name,
			Role:    cred.Field("Role"),
			Email:   cred.Field("Email"),
		})
	}

	geojson := ""
	if len(dataSheet.Rows) > 2 && len(dataSheet.Rows[2]) > 0 {
		geojson = strings.TrimSpace(dataSheet.Rows[2][0].Value)
	}

	page := MapPage{
		Slug:        slug,
		Section:     hdr.Field("Section"),
		Headline:    hdr.Field("Headline"),
		Eyebrow:     hdr.Field("Eyebrow"),
		Dek:         hdr.Field("Deck"),
		Byline:      hdr.Field("Author"),
		Date:        hdr.Field("Date"),
		Body:        hdr.Field("Introduction"),
		PublishedAt: publishedAt,
		Color:       set.Field("Map Color"),
		Blurb:       hdr.Field("Blurb"),
		Description: hdr.Field("Description"),
		InternalID:  hdr.Field("Internal ID"),
		Kicker:      hdr.Field("Kicker"),
		Topics:      topics,
		GeoJSON:     geojson,
		Credits:     credits,
	}

	return []MapPage{page}, nil
}
