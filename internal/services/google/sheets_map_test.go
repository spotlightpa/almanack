package google

import (
	"strings"
	"testing"
	"time"

	"github.com/earthboundkid/assert"
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
		assert.Continues(t).
			Equal(tooltipCards(tc.tooltipType, "desktop"), tc.desktop).
			Equal(tooltipCards(tc.tooltipType, "mobile"), tc.mobile)
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
		assert.Continues(t).Equal(sheetPercent(in), want)
	}
}

func TestSheetGravity(t *testing.T) {
	cases := map[string]string{
		"":         "ce",
		"Center":   "ce",
		"Top":      "no",
		" bottom ": "so",
		"Left":     "we",
		"Right":    "ea",
		"Smart":    "sm",
		"other":    "ce",
	}
	for in, want := range cases {
		assert.Continues(t).Equal(sheetGravity(in), want)
	}
}

func TestTemplateAttr(t *testing.T) {
	assert.Continues(t).
		Equal(templateAttr("\na\nb\n"), "a<br>b").
		Equal(templateAttr(`a\nb`), "a<br>b").
		Equal(templateAttr(`<div class="x">`), "<div class='x'>")
}

func TestMapContentPath(t *testing.T) {
	published := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	assert.Continues(t).
		Equal(mapContentPath("", "reading-map", "splreadingmap", published),
			"content/news/2026-07-20-SPLREADINGMAP.md").
		Equal(mapContentPath("News", "reading-map", "", time.Time{}),
			"content/news/reading-map.md")
}

func TestMapPageToMarkdown(t *testing.T) {
	be := assert.FailsNow(t)
	page := MapPage{
		Slug:            "reading-map",
		Headline:        "Find Your District",
		Byline:          "Hanna Holthaus",
		PublishedAt:     time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		TooltipsEnabled: true,
		TooltipValue:    "<b>{{Member}}</b>\n{{column}}\n<i>{{Party}}</i>",
		DesktopCards:    true,
		GeoJSON:         "https://example.com/map.geojson",
		ShareImage:      "2026/07/share.png",
		ShareImageDesc:  "Map of Reading council districts",
		ShareImageCrop:  "Top",
	}
	md := be.OK(page.ToMarkdown(""))
	be.True(strings.Contains(md, `tooltips="true"`)).
		True(strings.Contains(md, `desktop-display="card"`)).
		False(strings.Contains(md, "mobile-display")).
		False(strings.Contains(md, "eyebrow=")).
		True(strings.Contains(md, "{{column}}")).
		True(strings.Contains(md, `geojson="https://example.com/map.geojson"`)).
		True(strings.Contains(md, `image = "2026/07/share.png"`)).
		True(strings.Contains(md, `image-description = "Map of Reading council districts"`)).
		True(strings.Contains(md, `image-gravity = "no"`))
}
