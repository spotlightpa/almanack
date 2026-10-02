package google

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/utils/shortcode"
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
	return mapContentPath(m.Section, m.Slug, m.InternalID, m.PublishedAt)
}

func mapContentPath(section, slug, internalID string, publishedAt time.Time) string {
	if section == "" {
		section = "news"
	}
	section = strings.ToLower(section)

	name := slug
	if internalID != "" && !publishedAt.IsZero() {
		name = fmt.Sprintf("%s-%s",
			publishedAt.Format("2006-01-02"),
			strings.ToUpper(internalID),
		)
	}

	return fmt.Sprintf("content/%s/%s.md", section, name)
}

func (m MapPage) ToMarkdown(featuredMD string) (string, error) {
	var authors []string
	if m.Byline != "" {
		authors = stringx.ExtractNames(m.Byline)
	}
	fm, err := stringx.ToToml(map[string]any{
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
	sb.WriteString("+++\n")
	sb.WriteString(fm)
	sb.WriteString("+++\n")
	sb.WriteString("\n")

	var attrs []string
	if m.Eyebrow != "" {
		attrs = append(attrs, "eyebrow", m.Eyebrow)
	}
	if m.Headline != "" {
		attrs = append(attrs, "hed", m.Headline)
	}
	if m.Dek != "" {
		attrs = append(attrs, "dek", m.Dek)
	}
	if m.Date != "" {
		attrs = append(attrs, "display-date", m.Date)
	}
	if m.Byline != "" {
		attrs = append(attrs, "byline", m.Byline)
	}
	if m.Color != "" {
		attrs = append(attrs, "color", m.Color)
	}
	if m.Layout != "" {
		attrs = append(attrs, "layout", m.Layout)
	}
	if m.MobileLayout != "" {
		attrs = append(attrs, "mobile-layout", m.MobileLayout)
	}
	if m.ColorOpacity != "" {
		attrs = append(attrs, "color-opacity", m.ColorOpacity)
	}
	if m.SearchEnabled {
		attrs = append(attrs, "search", "true")
	}
	if m.SearchText != "" {
		attrs = append(attrs, "search-text", m.SearchText)
	}
	if m.SearchUseLocation {
		attrs = append(attrs, "search-use-location", "true")
	}
	if m.ReadMoreEnabled {
		attrs = append(attrs, "read-more", "true")
	}
	if m.TooltipsEnabled {
		attrs = append(attrs, "tooltips", "true")
	}
	if m.TooltipValue != "" {
		tv := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>").Replace(m.TooltipValue)
		tv = strings.ReplaceAll(tv, `\n`, "<br>")
		tv = trimBR(tv)
		tv = strings.ReplaceAll(tv, `"`, "'")
		attrs = append(attrs, "tooltip-value", tv)
	}
	if m.CustomCodeURL != "" {
		attrs = append(attrs, "custom-code", m.CustomCodeURL)
	}
	attrs = append(attrs, "outlet", "Spotlight PA")
	if m.GeoJSON != "" {
		attrs = append(attrs, "geojson", m.GeoJSON)
	}
	sb.WriteString(shortcode.New("featured/map-header", attrs...))
	sb.WriteString("\n")
	if m.Body != "" {
		sb.WriteString(m.Body)
		sb.WriteString("\n")
	}
	sb.WriteString("{{</featured/map-header>}}\n\n")

	if featuredMD != "" {
		sb.WriteString(featuredMD)
		sb.WriteString("\n")
	}

	writeMapCredits(&sb, m.Credits)

	return sb.String(), nil
}

func writeMapCredits(sb *strings.Builder, credits []MapCredit) {
	if len(credits) == 0 {
		return
	}
	sb.WriteString(shortcode.New("featured/footer"))
	sb.WriteString("\n")
	for _, c := range credits {
		var cattrs []string
		if c.Eyebrow != "" {
			cattrs = append(cattrs, "eyebrow", c.Eyebrow)
		}
		if c.Name != "" {
			cattrs = append(cattrs, "name", c.Name)
		}
		if c.Role != "" {
			cattrs = append(cattrs, "role", c.Role)
		}
		if c.Email != "" {
			cattrs = append(cattrs, "email", c.Email)
		}
		sb.WriteString(shortcode.New("featured/credit", cattrs...))
		sb.WriteString("\n")
	}
	sb.WriteString("{{</featured/footer>}}\n")
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

func (sm *sheetMapSkipDescription) Rows() iter.Seq[int] {
	return func(yield func(int) bool) {
		for row := range sm.SheetMap.Rows() {
			if row == 1 {
				continue
			}
			if !yield(row) {
				return
			}
		}
	}
}

func hasRow(seq iter.Seq[int]) bool {
	for range seq {
		return true
	}
	return false
}

func sheetPublished(ctx context.Context, s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.Time{}, err
	}
	parsed, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		l := almlog.FromContext(ctx)
		l.ErrorContext(ctx, "sheetPublished: invalid Published date", "value", s, "err", err)
		return time.Now(), nil
	}
	return parsed, nil
}

func sheetTopics(s string) []string {
	var topics []string
	for t := range strings.SplitSeq(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			topics = append(topics, t)
		}
	}
	return topics
}

func sheetCredits(sheet *spreadsheet.Sheet) []MapCredit {
	var credits []MapCredit
	cred := newSheetMapSkipDescription(sheet)
	for range cred.Rows() {
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
	return credits
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

	if !hasRow(hdr.Rows()) {
		return nil, resperr.E{M: "No data rows in Header sheet"}
	}
	if !hasRow(set.Rows()) {
		return nil, resperr.E{M: "No data rows in Map Settings sheet"}
	}
	if !hasRow(dat.Rows()) {
		return nil, resperr.E{M: "No data rows in Map Data sheet"}
	}
	if !hasRow(tip.Rows()) {
		return nil, resperr.E{M: "No data rows in Map Tooltips sheet"}
	}

	slug := hdr.Field("Slug")
	if slug == "" {
		return nil, resperr.E{M: "Header sheet missing Slug value"}
	}

	publishedAt, err := sheetPublished(ctx, hdr.Field("Published"))
	if err != nil {
		return nil, err
	}
	topics := sheetTopics(hdr.Field("Topics"))
	credits := sheetCredits(creditsSheet)

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
