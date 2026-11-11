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

func TestMapPageDefaultShareImage(t *testing.T) {
	be := assert.FailsNow(t)
	page := MapPage{Slug: "reading-map", Headline: "Find Your District"}
	md := be.OK(page.ToMarkdown(""))
	be.True(strings.Contains(md, `image = "2020/06/01fdjfj1nf20rwxt.jpeg"`)).
		True(strings.Contains(md, `image-description = "Spotlight PA"`))
}

func TestMapPageToMarkdown(t *testing.T) {
	be := assert.FailsNow(t)
	page := MapPage{
		Slug:             "reading-map",
		Headline:         "Find Your District",
		Byline:           "Hanna Holthaus",
		PublishedAt:      time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		TooltipsEnabled:  true,
		TooltipValue:     "<b>{{Member}}</b>\n{{column}}\n<i>{{Party}}</i>",
		DesktopCards:     true,
		GeoJSON:          "https://example.com/map.geojson",
		ShareImage:       "2026/07/share.png",
		ShareImageDesc:   "Map of Reading council districts",
		ShareImageCredit: "Reporter / Spotlight PA",
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
		True(strings.Contains(md, `image-gravity = "ce"`)).
		True(strings.Contains(md, `image-credit = "Jane Doe / Spotlight PA"`))
	be.Equal(page.ShareImagePath(), "2026/07/share.png")
	be.Equal(MapPage{}.ShareImagePath(), defaultShareImage)
}

func TestSheetPublished(t *testing.T) {
	be := assert.FailsNow(t)
	got := be.OK(sheetPublished("2026-07-20"))
	be.Equal(got.Format("2006-01-02 MST"), "2026-07-20 EDT")
	_, err := sheetPublished("")
	be.Nonzero(err)
	_, err = sheetPublished("July 20")
	be.Nonzero(err)
}

func TestSheetTopics(t *testing.T) {
	be := assert.Continues(t)
	be.SlicesEqual(sheetTopics("Local Government, Berks County ,,"), []string{"Local Government", "Berks County"})
	be.Equal(len(sheetTopics("")), 0)
}

func TestMapPageCredits(t *testing.T) {
	be := assert.FailsNow(t)
	page := MapPage{
		Slug: "reading-map",
		Credits: []MapCredit{
			{Eyebrow: "Reporting", Name: "Hanna Holthaus", Email: "hanna@spotlightpa.org"},
			{Name: "Alexandra Harris", Role: "Developer"},
		},
	}
	md := be.OK(page.ToMarkdown("Story text"))
	be.True(strings.Contains(md, "Story text\n{{<featured/footer>}}")).
		True(strings.Contains(md, `{{<featured/credit eyebrow="Reporting" name="Hanna Holthaus" email="hanna@spotlightpa.org">}}`)).
		True(strings.Contains(md, `{{<featured/credit name="Alexandra Harris" role="Developer">}}`)).
		True(strings.HasSuffix(md, "{{</featured/footer>}}\n"))
	be.False(strings.Contains(be.OK(MapPage{Slug: "x"}.ToMarkdown("")), "featured/footer"))
}
