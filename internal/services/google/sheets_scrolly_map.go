package google

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/almanack/internal/utils/shortcode"
	"github.com/spotlightpa/almanack/internal/utils/stringx"
	spreadsheet "gopkg.in/Iwark/spreadsheet.v2"
)

type ScrollyMapPage struct {
	Slug                   string
	Section                string
	Headline               string
	Eyebrow                string
	Dek                    string
	Byline                 string
	Date                   string
	PublishedAt            time.Time
	InternalID             string
	Kicker                 string
	Topics                 []string
	Blurb                  string
	Description            string
	ShareImage             string
	ShareImageDescription  string
	ShareImageCredit       string
	TopperLayout           string
	TopperMobileLayout     string
	TopperImage            string
	TopperImageDescription string
	Color                  string
	ColorOpacity           string
	CardPosition           string
	BasemapHidden          bool
	BasemapThrough         bool
	Background             string
	OutlineOff             bool
	OutlineColor           string
	ShadeProperty          string
	FeaturedDocLink        string
	FeaturedStoryPosition  string
	FeaturedStoryStop      string
	Stops                  []ScrollyMapStop
	Credits                []MapCredit
}

type ScrollyMapStop struct {
	SheetRow         int
	Label            string
	GeoJSON          string
	CardText         string
	Color            string
	CardPosition     string
	ZoomTo           string
	Dots             string
	Image            string
	ImageDescription string
	ImageCaption     string
	ImageCredit      string
	ImageRatio       string
	Datawrapper      string
	Flourish         string
	MediaPosition    string
	CardClass        string
	TextClass        string
}

func (m ScrollyMapPage) FilePath() string {
	return mapContentPath(m.Section, m.Slug, m.InternalID, m.PublishedAt)
}

func (m ScrollyMapPage) storySplit(featuredMD string) int {
	if featuredMD == "" {
		return -1
	}
	switch sheetOption(m.FeaturedStoryPosition) {
	case "before-the-map":
		return 0
	case "after-a-specific-stop":
		row, err := strconv.Atoi(strings.TrimLeft(strings.ToUpper(strings.TrimSpace(m.FeaturedStoryStop)), "ABCDEFGHIJKLMNOPQRSTUVWXYZ"))
		if err != nil {
			return -1
		}
		for i, stop := range m.Stops[:max(len(m.Stops)-1, 0)] {
			if stop.SheetRow == row {
				return i + 1
			}
		}
	}
	return -1
}

func (s ScrollyMapStop) hasContent() bool {
	return s.Label != "" || s.CardText != "" || s.GeoJSON != "" || s.Dots != "" ||
		s.Image != "" || s.Datawrapper != "" || s.Flourish != ""
}

func writeScrollyMap(sb *strings.Builder, attrs []string, stops []ScrollyMapStop) {
	sb.WriteString(shortcode.New("featured/scrolly-map", attrs...))
	sb.WriteString("\n")
	for _, stop := range stops {
		var sattrs []string
		sattrs = appendAttr(sattrs, "label", stop.Label)
		sattrs = appendAttr(sattrs, "color", stop.Color)
		sattrs = appendAttr(sattrs, "geojson", stop.GeoJSON)
		sattrs = appendAttr(sattrs, "align", sheetOption(stop.CardPosition))
		sattrs = appendAttr(sattrs, "fit", sheetOption(stop.ZoomTo))
		sattrs = appendAttr(sattrs, "dots", stop.Dots)
		sattrs = appendAttr(sattrs, "image", stop.Image)
		sattrs = appendAttr(sattrs, "image-description", stop.ImageDescription)
		sattrs = appendAttr(sattrs, "image-caption", stop.ImageCaption)
		sattrs = appendAttr(sattrs, "image-credit", stop.ImageCredit)
		sattrs = appendAttr(sattrs, "image-ratio", sheetRatio(stop.ImageRatio))
		sattrs = appendAttr(sattrs, "datawrapper", stop.Datawrapper)
		sattrs = appendAttr(sattrs, "flourish", stop.Flourish)
		sattrs = appendAttr(sattrs, "media-position", sheetOption(stop.MediaPosition))
		sattrs = appendAttr(sattrs, "card-class", stop.CardClass)
		sattrs = appendAttr(sattrs, "text-class", stop.TextClass)
		sb.WriteString(shortcode.New("featured/scrolly-map-stop", sattrs...))
		sb.WriteString("\n")
		if stop.CardText != "" {
			sb.WriteString(stop.CardText)
			sb.WriteString("\n")
		}
		sb.WriteString("{{</featured/scrolly-map-stop>}}\n")
	}
	sb.WriteString("{{</featured/scrolly-map>}}\n")
}

func (m ScrollyMapPage) share() (image, description, gravity string) {
	if m.ShareImage != "" {
		return m.ShareImage, m.ShareImageDescription, "ce"
	}
	if m.TopperImage != "" {
		if strings.HasPrefix(sheetOption(m.TopperLayout), "icon") {
			return m.TopperImage, m.TopperImageDescription, "ce"
		}
		return m.TopperImage, m.TopperImageDescription, "sm"
	}
	return defaultShareImage, defaultShareImageDescription, "ce"
}

func (m ScrollyMapPage) ShareImagePath() string {
	image, _, _ := m.share()
	return image
}

func (m ScrollyMapPage) ToMarkdown(featuredMD string) (string, error) {
	var authors []string
	if m.Byline != "" {
		authors = stringx.ExtractNames(m.Byline)
	}
	shareImage, shareDescription, shareGravity := m.share()
	fm, err := stringx.ToToml(map[string]any{
		"authors":           authors,
		"blurb":             m.Blurb,
		"byline":            m.Byline,
		"description":       m.Description,
		"image":             shareImage,
		"image-credit":      m.ShareImageCredit,
		"image-description": shareDescription,
		"image-gravity":     shareGravity,
		"internal-id":       m.InternalID,
		"kicker":            m.Kicker,
		"layout":            "scrolly-map",
		"published":         m.PublishedAt,
		"slug":              m.Slug,
		"suppress-ads":      true,
		"title":             m.Headline,
		"title-tag":         m.Headline,
		"topics":            m.Topics,
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
	attrs = appendAttr(attrs, "eyebrow", m.Eyebrow)
	attrs = appendAttr(attrs, "hed", m.Headline)
	attrs = appendAttr(attrs, "dek", m.Dek)
	attrs = appendAttr(attrs, "display-date", m.Date)
	attrs = appendAttr(attrs, "byline", m.Byline)
	attrs = append(attrs, "outlet", "Spotlight PA")
	attrs = appendAttr(attrs, "color", m.Color)
	attrs = appendAttr(attrs, "color-opacity", m.ColorOpacity)
	attrs = appendAttr(attrs, "layout", sheetOption(m.CardPosition))
	if m.BasemapHidden {
		attrs = append(attrs, "basemap", "hide")
		attrs = appendAttr(attrs, "background", m.Background)
	} else if m.BasemapThrough {
		attrs = append(attrs, "basemap", "show-through")
	}
	if m.OutlineOff {
		attrs = append(attrs, "outline", "false")
	}
	attrs = appendAttr(attrs, "outline-color", m.OutlineColor)
	attrs = appendAttr(attrs, "shade-property", m.ShadeProperty)
	if layout := sheetOption(m.TopperLayout); m.TopperImage != "" && layout != "none" {
		attrs = append(attrs, "topper-image", m.TopperImage)
		attrs = appendAttr(attrs, "topper-image-description", m.TopperImageDescription)
		attrs = appendAttr(attrs, "topper-layout", layout)
		attrs = appendAttr(attrs, "topper-mobile-layout", sheetOption(m.TopperMobileLayout))
	}
	split := m.storySplit(featuredMD)
	if split < 0 {
		writeScrollyMap(&sb, attrs, m.Stops)
		sb.WriteString("\n")
		if featuredMD != "" {
			sb.WriteString(featuredMD)
			sb.WriteString("\n")
		}
	} else {
		writeScrollyMap(&sb, append(slices.Clip(attrs), "continues", "true"), m.Stops[:split])
		sb.WriteString("\n")
		sb.WriteString(featuredMD)
		sb.WriteString("\n\n")
		writeScrollyMap(&sb, append(slices.Clip(attrs), "continued", "true"), m.Stops[split:])
		sb.WriteString("\n")
	}

	writeMapCredits(&sb, m.Credits)

	return sb.String(), nil
}

func appendAttr(attrs []string, key, value string) []string {
	if value == "" {
		return attrs
	}
	return append(attrs, key, value)
}

func sheetOption(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), "-")
}

func sheetUnchecked(s string) bool {
	return strings.EqualFold(strings.TrimSpace(s), "FALSE")
}

func SizeRatio(w, h int) string {
	if w <= 0 || h <= 0 {
		return ""
	}
	a, b := w, h
	for b != 0 {
		a, b = b, a%b
	}
	return fmt.Sprintf("%d:%d", w/a, h/a)
}

func sheetRatio(s string) string {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return ""
	}
	w, werr := strconv.Atoi(parts[0])
	h, herr := strconv.Atoi(parts[1])
	if werr != nil || herr != nil || w <= 0 || h <= 0 {
		return ""
	}
	return fmt.Sprintf("%d:%d", w, h)
}

func sheetText(s string) string {
	s = strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(s)
	return strings.TrimSpace(s)
}

func SheetToScrollyMapPages(ctx context.Context, cl *http.Client, sheetID string) ([]ScrollyMapPage, error) {
	service := spreadsheet.NewServiceWithClient(cl)
	doc, err := service.FetchSpreadsheet(sheetID)
	if err != nil {
		return nil, resperr.E{E: err, M: "Problem fetching scrolly map config sheet"}
	}

	headerSheet, err := doc.SheetByTitle("Header")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Header' sheet"}
	}

	settingsSheet, err := doc.SheetByTitle("Map Settings")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Map Settings' sheet"}
	}

	stopsSheet, err := doc.SheetByTitle("Stops")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Stops' sheet"}
	}

	creditsSheet, err := doc.SheetByTitle("Credits")
	if err != nil {
		return nil, resperr.E{E: err, M: "Spreadsheet missing 'Credits' sheet"}
	}

	hdr := newSheetMapSkipDescription(headerSheet)
	set := newSheetMapSkipDescription(settingsSheet)

	if !hasRow(hdr.Rows()) {
		return nil, resperr.E{M: "No data rows in Header sheet"}
	}
	if !hasRow(set.Rows()) {
		return nil, resperr.E{M: "No data rows in Map Settings sheet"}
	}

	slug := hdr.Field("Slug")
	if slug == "" {
		return nil, resperr.E{M: "Header sheet missing Slug value"}
	}

	publishedAt, err := sheetPublished(ctx, hdr.Field("Published"))
	if err != nil {
		return nil, err
	}

	var stops []ScrollyMapStop
	st := newSheetMapSkipDescription(stopsSheet)
	for row := range st.Rows() {
		stop := ScrollyMapStop{
			SheetRow:         row + 1,
			Label:            st.Field("Label"),
			GeoJSON:          st.Field("GeoJSON"),
			CardText:         sheetText(st.Field("Card Text")),
			Color:            st.Field("Color"),
			CardPosition:     st.Field("Card Position"),
			ZoomTo:           st.Field("Zoom To"),
			Dots:             st.Field("Dots"),
			Image:            st.Field("Image"),
			ImageDescription: st.Field("Image Description"),
			ImageCaption:     st.Field("Image Caption"),
			ImageCredit:      st.Field("Image Credit"),
			ImageRatio:       st.Field("Image Ratio"),
			Datawrapper:      st.Field("Datawrapper"),
			Flourish:         st.Field("Flourish"),
			MediaPosition:    st.Field("Media Position"),
			CardClass:        st.Field("Card Class"),
			TextClass:        st.Field("Text Class"),
		}
		if !stop.hasContent() {
			continue
		}
		stops = append(stops, stop)
	}
	if len(stops) == 0 {
		return nil, resperr.E{M: "No stops in Stops sheet"}
	}

	page := ScrollyMapPage{
		Slug:                   slug,
		Section:                hdr.Field("Section"),
		Headline:               hdr.Field("Headline"),
		Eyebrow:                hdr.Field("Eyebrow"),
		Dek:                    hdr.Field("Deck"),
		Byline:                 hdr.Field("Author"),
		Date:                   hdr.Field("Display Date"),
		PublishedAt:            publishedAt,
		InternalID:             hdr.Field("Internal ID"),
		Kicker:                 hdr.Field("Kicker"),
		Topics:                 sheetTopics(hdr.Field("Topics")),
		Blurb:                  hdr.Field("Blurb"),
		Description:            hdr.Field("Description"),
		ShareImage:             hdr.Field("Share Image"),
		ShareImageDescription:  hdr.Field("Share Image Description"),
		TopperLayout:           hdr.Field("Topper Layout"),
		TopperMobileLayout:     hdr.Field("Mobile Topper Layout"),
		TopperImage:            hdr.Field("Topper Image"),
		TopperImageDescription: hdr.Field("Topper Image Description"),
		Color:                  set.Field("Map Color"),
		ColorOpacity:           sheetPercent(set.Field("Map Color Opacity")),
		CardPosition:           set.Field("Default Card Position"),
		BasemapHidden:          strings.EqualFold(set.Field("Basemap"), "Hide"),
		BasemapThrough:         sheetOption(set.Field("Basemap")) == "show-through",
		Background:             set.Field("Background Color"),
		OutlineOff:             sheetUnchecked(set.Field("State Outline")),
		OutlineColor:           set.Field("Outline Color"),
		ShadeProperty:          set.Field("Shade Property"),
		FeaturedDocLink:        set.Field("Featured Story Document Link"),
		FeaturedStoryPosition:  set.Field("Featured Story Position"),
		FeaturedStoryStop:      set.Field("Featured Story Stop"),
		Stops:                  stops,
		Credits:                sheetCredits(creditsSheet),
	}

	return []ScrollyMapPage{page}, nil
}
