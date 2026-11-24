package google

import (
	"cmp"
	"context"
	"iter"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/earthboundkid/bytemap/v2"
	"github.com/earthboundkid/resperr/v2"
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
	SearchEnabled     bool
	SearchText        string
	SearchUseLocation bool
	ReadMoreEnabled   bool
	TooltipsEnabled   bool
	TooltipValue      string
	DesktopCards      bool
	MobileCards       bool
	FeaturedDocLink   string
	CustomCodeURL     string
	Blurb             string
	Description       string
	ShareImage        string
	ShareImageDesc    string
	ShareImageCredit  string
	InternalID        string
	Kicker            string
	Topics            []string
	Intro             string
	Layout            string
	MobileLayout      string
	Credits           []MapCredit
}

type MapSheet interface {
	Frontmatter() map[string]any
	Body(featuredMD string) string
	FeaturedDoc() string
	PageKind() string
	ShareImagePath() string
}

func SheetToMap(ctx context.Context, cl *http.Client, sheetID string) (MapSheet, error) {
	service := spreadsheet.NewServiceWithClient(cl)
	doc, err := service.FetchSpreadsheet(sheetID)
	if err != nil {
		return nil, resperr.E{E: err, M: "Problem fetching map sheet"}
	}
	if _, err := doc.SheetByTitle("Stops"); err == nil {
		return sheetToScrollyMapPage(&doc)
	}
	return sheetToMapPage(&doc)
}

var validSheetIDChars = bytemap.Union(
	bytemap.Range('A', 'Z'),
	bytemap.Range('a', 'z'),
	bytemap.Range('0', '9'),
	bytemap.Make("-_"),
)

func NormalizeSheetID(id string) (string, error) {
	const magicLength = 44
	var v resperr.Validator
	v.AddIf("sheet_id", id == "", "ID must be set")
	if after, ok := strings.CutPrefix(id, "https://docs.google.com/spreadsheets/d/"); ok {
		id, _, _ = strings.Cut(after, "/")
	}
	v.AddIfUnset("sheet_id", len(id) != magicLength, "Invalid Google Sheet ID or link")
	v.AddIfUnset("sheet_id", !validSheetIDChars.Contains(id), "Illegal characters in Google Sheet ID")
	if err := v.Err(); err != nil {
		return "", err
	}
	return id, nil
}

func sheetSection(s string) string {
	return cmp.Or(strings.ToLower(strings.TrimSpace(s)), "news")
}

type MapCredit struct {
	Eyebrow string
	Name    string
	Role    string
	Email   string
}

const (
	defaultShareImage            = "2020/06/01fdjfj1nf20rwxt.jpeg"
	defaultShareImageDescription = "Spotlight PA"
)

func (m *MapPage) share() (image, description string) {
	if m.ShareImage == "" {
		return defaultShareImage, defaultShareImageDescription
	}
	return m.ShareImage, m.ShareImageDesc
}

func (m *MapPage) ShareImagePath() string {
	image, _ := m.share()
	return image
}

func (m *MapPage) Frontmatter() map[string]any {
	shareImage, shareDescription := m.share()
	var authors []string
	if m.Byline != "" {
		authors = stringx.ExtractNames(m.Byline)
	}
	return map[string]any{
		"authors":           authors,
		"blurb":             m.Blurb,
		"byline":            m.Byline,
		"description":       m.Description,
		"image":             shareImage,
		"image-credit":      m.ShareImageCredit,
		"image-description": shareDescription,
		"image-gravity":     "ce",
		"internal-id":       m.InternalID,
		"kicker":            m.Kicker,
		"layout":            "searchable-map",
		"published":         m.PublishedAt,
		"slug":              m.Slug,
		"suppress-ads":      true,
		"title":             m.Headline,
		"title-tag":         m.Headline,
		"topics":            m.Topics,
	}
}

func (m *MapPage) Body(featuredMD string) string {
	var sb strings.Builder

	var attrs []string
	attrs = appendAttr(attrs, "eyebrow", m.Eyebrow)
	attrs = appendAttr(attrs, "hed", m.Headline)
	attrs = appendAttr(attrs, "dek", m.Dek)
	attrs = appendAttr(attrs, "display-date", m.Date)
	attrs = appendAttr(attrs, "byline", m.Byline)
	attrs = appendAttr(attrs, "color", m.Color)
	attrs = appendAttr(attrs, "layout", m.Layout)
	attrs = appendAttr(attrs, "mobile-layout", m.MobileLayout)
	attrs = appendAttr(attrs, "color-opacity", m.ColorOpacity)
	if m.SearchEnabled {
		attrs = append(attrs, "search", "true")
	}
	attrs = appendAttr(attrs, "search-text", m.SearchText)
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
		attrs = append(attrs, "tooltip-value", templateAttr(m.TooltipValue))
	}
	if m.DesktopCards {
		attrs = append(attrs, "desktop-display", "card")
	}
	if m.MobileCards {
		attrs = append(attrs, "mobile-display", "card")
	}
	attrs = appendAttr(attrs, "custom-code", m.CustomCodeURL)
	attrs = append(attrs, "outlet", "Spotlight PA")
	attrs = appendAttr(attrs, "geojson", m.GeoJSON)
	sb.WriteString(shortcode.New("featured/map-header", attrs...))
	sb.WriteString("\n")
	if m.Intro != "" {
		sb.WriteString(m.Intro)
		sb.WriteString("\n")
	}
	sb.WriteString("{{</featured/map-header>}}\n\n")

	if featuredMD != "" {
		sb.WriteString(featuredMD)
		sb.WriteString("\n")
	}

	writeMapCredits(&sb, m.Credits, featuredMD != "")

	return sb.String()
}

func (m *MapPage) FeaturedDoc() string {
	return m.FeaturedDocLink
}

func (m *MapPage) PageKind() string {
	return sheetSection(m.Section)
}

func writeMapCredits(sb *strings.Builder, credits []MapCredit, storyAbove bool) {
	if len(credits) == 0 && !storyAbove {
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

func templateAttr(s string) string {
	v := strings.NewReplacer("\r\n", "<br>", "\n", "<br>", "\r", "<br>").Replace(s)
	v = strings.ReplaceAll(v, `\n`, "<br>")
	v = trimBR(v)
	return strings.ReplaceAll(v, `"`, "'")
}

func tooltipCards(tooltipType, screen string) bool {
	t := strings.ToLower(tooltipType)
	if !strings.Contains(t, "card") {
		return false
	}
	desktop := strings.Contains(t, "desktop")
	mobile := strings.Contains(t, "mobile")
	if !desktop && !mobile {
		return true
	}
	return (screen == "desktop" && desktop) || (screen == "mobile" && mobile)
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

func sheetPublished(s string) (time.Time, error) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.Time{}, err
	}
	parsed, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		return time.Time{}, resperr.E{E: err, M: "Header sheet needs a Published date in YYYY-MM-DD format"}
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

func sheetToMapPage(doc *spreadsheet.Spreadsheet) (*MapPage, error) {
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

	publishedAt, err := sheetPublished(hdr.Field("Published"))
	if err != nil {
		return nil, err
	}
	topics := sheetTopics(hdr.Field("Topics"))
	credits := sheetCredits(creditsSheet)

	geojson := dat.Field("Map Data")

	page := &MapPage{
		Slug:              slug,
		Section:           hdr.Field("Section"),
		Headline:          hdr.Field("Headline"),
		Eyebrow:           hdr.Field("Eyebrow"),
		Dek:               hdr.Field("Deck"),
		Byline:            hdr.Field("Author"),
		Date:              hdr.Field("Display Date"),
		Layout:            hdr.Field("Map Layout"),
		MobileLayout:      hdr.Field("Mobile Map Layout"),
		Intro:             hdr.Field("Introduction"),
		PublishedAt:       publishedAt,
		Color:             set.Field("Map Color"),
		ColorOpacity:      sheetPercent(set.Field("Map Color Opacity")),
		SearchEnabled:     sheetBool(set.Field("Search Bar")),
		SearchText:        set.Field("Search Bar Text"),
		SearchUseLocation: sheetBool(set.Field("Search Bar Use Location")),
		ReadMoreEnabled:   sheetBool(set.Field("Read More")),
		TooltipsEnabled:   sheetBool(tip.Field("Tooltips Enabled")),
		TooltipValue:      tip.Field("Tooltip Value"),
		DesktopCards:      tooltipCards(tip.Field("Tooltip Type"), "desktop"),
		MobileCards:       tooltipCards(tip.Field("Tooltip Type"), "mobile"),
		FeaturedDocLink:   set.Field("Featured Story Document Link"),
		CustomCodeURL:     set.Field("Custom Code"),
		Blurb:             hdr.Field("Blurb"),
		Description:       hdr.Field("Description"),
		ShareImage:        hdr.Field("Share Image"),
		ShareImageDesc:    hdr.Field("Share Image Description"),
		InternalID:        hdr.Field("Internal ID"),
		Kicker:            hdr.Field("Kicker"),
		Topics:            topics,
		GeoJSON:           geojson,
		Credits:           credits,
	}

	return page, nil
}
