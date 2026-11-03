package google

import (
	"strings"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
)

func TestTooltipCards(t *testing.T) {
	cases := []struct {
		tooltipType     string
		desktop, mobile bool
	}{
		{"", false, false},
		{"Bubbles", false, false},
		{"Cards", true, true},
		{"cards", true, true},
		{"Cards (desktop only)", true, false},
		{"Cards (mobile only)", false, true},
	}
	for _, tc := range cases {
		be.Equal(t, tc.desktop, tooltipCards(tc.tooltipType, "desktop"))
		be.Equal(t, tc.mobile, tooltipCards(tc.tooltipType, "mobile"))
	}
}

func TestSheetPercent(t *testing.T) {
	cases := map[string]string{
		"":     "",
		"20%":  "0.2",
		"0.2":  "0.2",
		"20":   "0.2",
		"100%": "1",
		"abc":  "",
	}
	for in, want := range cases {
		be.Equal(t, want, sheetPercent(in))
	}
}

func TestTemplateAttr(t *testing.T) {
	be.Equal(t, "a<br>b", templateAttr("\na\nb\n"))
	be.Equal(t, "a<br>b", templateAttr(`a\nb`))
	be.Equal(t, "<div class='x'>", templateAttr(`<div class="x">`))
}

func TestMapContentPath(t *testing.T) {
	published := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	be.Equal(t, "content/news/2026-07-20-SPLREADINGMAP.md",
		mapContentPath("", "reading-map", "splreadingmap", published))
	be.Equal(t, "content/news/reading-map.md",
		mapContentPath("News", "reading-map", "", time.Time{}))
}

func TestMapPageToMarkdown(t *testing.T) {
	page := MapPage{
		Slug:            "reading-map",
		Headline:        "Find Your District",
		Byline:          "Hanna Holthaus",
		PublishedAt:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		TooltipsEnabled: true,
		TooltipValue:    "<b>{{Member}}</b>\n{{column}}\n<i>{{Party}}</i>",
		DesktopCards:    true,
		GeoJSON:         "https://example.com/map.geojson",
	}
	md, err := page.ToMarkdown("")
	be.NilErr(t, err)
	be.In(t, `tooltips="true"`, md)
	be.In(t, `desktop-display="card"`, md)
	be.True(t, !strings.Contains(md, "mobile-display"))
	be.True(t, !strings.Contains(md, "eyebrow="))
	be.In(t, "{{column}}", md)
	be.In(t, `geojson="https://example.com/map.geojson"`, md)
}
