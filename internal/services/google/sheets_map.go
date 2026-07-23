package google

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/utils/stringx"
	spreadsheet "gopkg.in/Iwark/spreadsheet.v2"
)

type MapPage struct {
	Slug              string
	Section           string
	Headline          string
	Eyebrow           string
	Dek               string
	Byline            string
	Date              string
	PublishedAt       time.Time
	GeoJSON           string
	Color             string
	ColorOpacity      string
	MapType           string
	SearchEnabled     bool
	SearchText        string
	SearchUseLocation bool
	ReadMoreEnabled   bool
	TooltipsEnabled   bool
	TooltipValue      string
	FeaturedDocLink   string
	CustomCodeURL     string
	Blurb             string
	Description       string
	InternalID        string
	Kicker            string
	Topics            []string
	Body              string
	Layout            string
	MobileLayout      string
	Credits           []MapCredit
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
	section = strings.ToLower(section)

	name := m.Slug
	if m.InternalID != "" && !m.PublishedAt.IsZero() {
		name = fmt.Sprintf("%s-%s",
			m.PublishedAt.Format("2006-01-02"),
			strings.ToUpper(m.InternalID),
		)
	}

	return fmt.Sprintf("content/%s/%s.md", section, name)
}

func (m MapPage) ToMarkdown(featuredMD string) (string, error) {
	var authors []string
	if m.Byline != "" {
		authors = stringx.ExtractNames(m.Byline)
	}
	fm, err := db.FrontmatterTOML(map[string]any{
		"authors":      authors,
		"blurb":        m.Blurb,
		"byline":       m.Byline,
		"description":  m.Description,
		"internal-id":  m.InternalID,
		"kicker":       m.Kicker,
		"layout":       "searchable-map",
		"published":    m.PublishedAt,
		"slug":         m.Slug,
		"suppress-ads": true,
		"title":        m.Headline,
		"title-tag":    m.Headline,
		"topics":       m.Topics,
	})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fm)
	sb.WriteString("\n")

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
		fmt.Fprintf(&sb, "  display-date=%q\n", m.Date)
	}
	if m.Byline != "" {
		fmt.Fprintf(&sb, "  byline=%q\n", m.Byline)
	}
	if m.Color != "" {
		fmt.Fprintf(&sb, "  color=%q\n", m.Color)
	}
	if m.Layout != "" {
		fmt.Fprintf(&sb, "  layout=%q\n", m.Layout)
	}
	if m.MobileLayout != "" {
		fmt.Fprintf(&sb, "  mobile-layout=%q\n", m.MobileLayout)
	}
	if m.ColorOpacity != "" {
		fmt.Fprintf(&sb, "  color-opacity=%q\n", m.ColorOpacity)
	}
	if m.SearchEnabled {
		sb.WriteString("  search=\"true\"\n")
	}
	if m.SearchText != "" {
		fmt.Fprintf(&sb, "  search-text=%q\n", m.SearchText)
	}
	if m.SearchUseLocation {
		sb.WriteString("  search-use-location=\"true\"\n")
	}
	if m.ReadMoreEnabled {
		sb.WriteString("  read-more=\"true\"\n")
	}
	if m.TooltipsEnabled {
		sb.WriteString("  tooltips=\"true\"\n")
	}
	if m.TooltipValue != "" {
		tv := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>").Replace(m.TooltipValue)
		tv = strings.ReplaceAll(tv, `\n`, "<br>")
		tv = trimBR(tv)
		tv = strings.ReplaceAll(tv, `"`, "'")
		fmt.Fprintf(&sb, "  tooltip-value=%q\n", tv)
	}
	if m.CustomCodeURL != "" {
		fmt.Fprintf(&sb, "  custom-code=%q\n", m.CustomCodeURL)
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

	if featuredMD != "" {
		sb.WriteString(featuredMD)
		sb.WriteString("\n")
	}

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

	return sb.String(), nil
}

func sheetBool(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "TRUE")
}

func trimBR(s string) string {
	for strings.HasPrefix(s, "<br>") {
		s = strings.TrimPrefix(s, "<br>")
	}
	for strings.HasSuffix(s, "<br>") {
		s = strings.TrimSuffix(s, "<br>")
	}
	return strings.TrimSpace(s)
}

func sheetPercent(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasSuffix(s, "%") {
		n, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(n/100, 'f', -1, 64)
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ""
	}
	if n > 1 {
		n = n / 100
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

type sheetMapSkipDescription struct {
	SheetMap
}

func newSheetMapSkipDescription(sheet *spreadsheet.Sheet) *sheetMapSkipDescription {
	return &sheetMapSkipDescription{SheetMap: *NewSheetMap(sheet)}
}

func (sm *sheetMapSkipDescription) Next() bool {
	if !sm.SheetMap.Next() {
		return false
	}
	if sm.row == 1 {
		return sm.SheetMap.Next()
	}
	return true
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

	tooltipSheet, err := doc.SheetByTitle("Map Tooltips")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Map Tooltips' sheet"}
	}

	hdr := newSheetMapSkipDescription(headerSheet)
	set := newSheetMapSkipDescription(settingsSheet)
	dat := newSheetMapSkipDescription(dataSheet)
	tip := newSheetMapSkipDescription(tooltipSheet)

	if !hdr.Next() {
		return nil, resperr.E{M: "No data rows in Header sheet"}
	}
	if !set.Next() {
		return nil, resperr.E{M: "No data rows in Map Settings sheet"}
	}
	if !dat.Next() {
		return nil, resperr.E{M: "No data rows in Map Data sheet"}
	}
	if !tip.Next() {
		return nil, resperr.E{M: "No data rows in Map Tooltips sheet"}
	}

	slug := hdr.Field("Slug")
	if slug == "" {
		return nil, resperr.E{M: "Header sheet missing Slug value"}
	}

	publishedStr := hdr.Field("Published")
	publishedAt := time.Now()
	if publishedStr != "" {
		loc, lerr := time.LoadLocation("America/New_York")
		if lerr != nil {
			return nil, lerr
		}
		parsed, perr := time.ParseInLocation("2006-01-02", publishedStr, loc)
		if perr != nil {
			l := almlog.FromContext(ctx)
			l.ErrorContext(ctx, "SheetToMapPages: invalid Published date", "value", publishedStr, "err", perr)
		} else {
			publishedAt = parsed
		}
	}

	topicsStr := hdr.Field("Topics")
	var topics []string
	for t := range strings.SplitSeq(topicsStr, ",") {
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

	geojson := dat.Field("Map Data")

	page := MapPage{
		Slug:              slug,
		Section:           hdr.Field("Section"),
		Headline:          hdr.Field("Headline"),
		Eyebrow:           hdr.Field("Eyebrow"),
		Dek:               hdr.Field("Deck"),
		Byline:            hdr.Field("Author"),
		Date:              hdr.Field("Display Date"),
		Layout:            hdr.Field("Map Layout"),
		MobileLayout:      hdr.Field("Mobile Map Layout"),
		Body:              hdr.Field("Introduction"),
		PublishedAt:       publishedAt,
		Color:             set.Field("Map Color"),
		ColorOpacity:      sheetPercent(set.Field("Map Color Opacity")),
		MapType:           set.Field("Map Type"),
		SearchEnabled:     sheetBool(set.Field("Search Bar")),
		SearchText:        set.Field("Search Bar Text"),
		SearchUseLocation: sheetBool(set.Field("Search Bar Use Location")),
		ReadMoreEnabled:   sheetBool(set.Field("Read More")),
		TooltipsEnabled:   sheetBool(tip.Field("Tooltips Enabled")),
		TooltipValue:      tip.Field("Tooltip Value"),
		FeaturedDocLink:   set.Field("Featured Story Document Link"),
		CustomCodeURL:     set.Field("Custom Code"),
		Blurb:             hdr.Field("Blurb"),
		Description:       hdr.Field("Description"),
		InternalID:        hdr.Field("Internal ID"),
		Kicker:            hdr.Field("Kicker"),
		Topics:            topics,
		GeoJSON:           geojson,
		Credits:           credits,
	}

	return []MapPage{page}, nil
}
