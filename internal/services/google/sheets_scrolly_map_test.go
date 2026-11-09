package google

import (
	"strings"
	"testing"
	"time"

	"github.com/earthboundkid/assert"
)

func TestSheetOption(t *testing.T) {
	assert.Continues(t).
		Equal(sheetOption("Icon Dark"), "icon-dark").
		Equal(sheetOption("Icon Light"), "icon-light").
		Equal(sheetOption("  Left "), "left").
		Equal(sheetOption(""), "")
}

func TestSheetRatio(t *testing.T) {
	cases := map[string]string{
		"16:9":  "16:9",
		" 3:2 ": "3:2",
		"4:3:0": "4:3",
		"16x9":  "",
		"0:9":   "",
		"a:b":   "",
		"":      "",
	}
	for in, want := range cases {
		assert.Continues(t).Equal(sheetRatio(in), want)
	}
}

func TestSheetUnchecked(t *testing.T) {
	assert.Continues(t).
		True(sheetUnchecked("FALSE")).
		True(sheetUnchecked(" false ")).
		False(sheetUnchecked("TRUE")).
		False(sheetUnchecked(""))
}

func TestSheetText(t *testing.T) {
	assert.Continues(t).
		Equal(sheetText("  a\r\nb\rc  "), "a\nb\nc")
}

func TestScrollyMapPageToMarkdown(t *testing.T) {
	be := assert.FailsNow(t)
	page := ScrollyMapPage{
		Slug:          "regions",
		Headline:      "Defining Pennsylvania's Regions",
		Byline:        "Asha Prihar",
		PublishedAt:   time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		TopperLayout:  "Icon Light",
		TopperImage:   "2026/09/topper.png",
		CardPosition:  "Right",
		BasemapHidden: true,
		Background:    "Navy",
		OutlineOff:    true,
		Stops: []ScrollyMapStop{
			{Label: "Central PA", GeoJSON: "https://example.com/a.geojson", CardText: "Hello", ZoomTo: "Shaded", ImageRatio: "16:9"},
		},
	}
	md := be.OK(page.ToMarkdown(""))
	for _, want := range []string{
		`image-gravity = "ce"`,
		`topper-layout="icon-light"`,
		`layout="right"`,
		`basemap="hide"`,
		`background="Navy"`,
		`outline="false"`,
		`fit="shaded"`,
		`image-ratio="16:9"`,
		"Hello\n{{</featured/scrolly-map-stop>}}",
	} {
		be.True(strings.Contains(md, want))
	}
}

func TestScrollyMapShareImageFallback(t *testing.T) {
	be := assert.FailsNow(t)
	page := ScrollyMapPage{
		Slug:                   "regions",
		Headline:               "Regions",
		TopperImage:            "2026/09/topper.png",
		TopperImageDescription: "Illustrated map of Pennsylvania",
		TopperLayout:           "Right",
		Stops:                  []ScrollyMapStop{{Label: "A"}},
	}
	md := be.OK(page.ToMarkdown(""))
	be.True(strings.Contains(md, `image = "2026/09/topper.png"`)).
		True(strings.Contains(md, `image-description = "Illustrated map of Pennsylvania"`)).
		True(strings.Contains(md, `image-gravity = "sm"`))

	page.TopperLayout = "Icon Light"
	md = be.OK(page.ToMarkdown(""))
	be.True(strings.Contains(md, `image-gravity = "ce"`))

	page.ShareImage = "2026/09/share.png"
	page.ShareImageDescription = "Share art"
	page.TopperLayout = "Right"
	md = be.OK(page.ToMarkdown(""))
	be.True(strings.Contains(md, `image = "2026/09/share.png"`)).
		True(strings.Contains(md, `image-description = "Share art"`)).
		True(strings.Contains(md, `image-gravity = "ce"`)).
		False(strings.Contains(md, `image = "2026/09/topper.png"`))
}

func TestScrollyMapTopperSkippedWithoutImage(t *testing.T) {
	be := assert.FailsNow(t)
	page := ScrollyMapPage{
		Slug:         "regions",
		Headline:     "Regions",
		TopperLayout: "Icon Light",
		Stops:        []ScrollyMapStop{{Label: "A"}},
	}
	md := be.OK(page.ToMarkdown(""))
	be.False(strings.Contains(md, "topper-layout")).
		False(strings.Contains(md, "topper-image")).
		True(strings.Contains(md, `image = "2020/06/01fdjfj1nf20rwxt.jpeg"`)).
		True(strings.Contains(md, `image-gravity = "ce"`))
}

func storyPage(position, stop string) ScrollyMapPage {
	return ScrollyMapPage{
		Slug:                  "regions",
		Headline:              "Regions",
		FeaturedStoryPosition: position,
		FeaturedStoryStop:     stop,
		Stops: []ScrollyMapStop{
			{SheetRow: 3, Label: "One"},
			{SheetRow: 4, Label: "Two"},
			{SheetRow: 6, Label: "Three"},
		},
	}
}

func TestScrollyMapStoryPosition(t *testing.T) {
	const story = "STORY TEXT"
	cases := []struct {
		position, stop string
		want           []string
	}{
		{"", "", []string{"One", "Two", "Three", "/featured/scrolly-map", story}},
		{"After the Map", "", []string{"One", "Two", "Three", "/featured/scrolly-map", story}},
		{"Before the Map", "", []string{`continues="true"`, "/featured/scrolly-map", story, `continued="true"`, "One", "Two", "Three"}},
		{"After a specific Stop", "A4", []string{`continues="true"`, "One", "Two", "/featured/scrolly-map", story, `continued="true"`, "Three"}},
		{"After a specific Stop", "3", []string{`continues="true"`, "One", "/featured/scrolly-map", story, `continued="true"`, "Two", "Three"}},
		{"After a specific Stop", "A6", []string{"One", "Two", "Three", "/featured/scrolly-map", story}},
		{"After a specific Stop", "A9", []string{"One", "Two", "Three", "/featured/scrolly-map", story}},
		{"After a specific Stop", "", []string{"One", "Two", "Three", "/featured/scrolly-map", story}},
	}
	for _, tc := range cases {
		md, err := storyPage(tc.position, tc.stop).ToMarkdown(story)
		assert.Continues(t).OK(md, err)
		pos := 0
		for _, want := range tc.want {
			i := strings.Index(md[pos:], want)
			assert.Continues(t).True(i >= 0)
			if i >= 0 {
				pos += i + len(want)
			}
		}
		splits := strings.Count(md, "{{</featured/scrolly-map>}}")
		wantSplits := 1
		if strings.Contains(tc.want[0], "continues") {
			wantSplits = 2
		}
		assert.Continues(t).Equal(splits, wantSplits)
	}
}

func TestScrollyMapStoryPositionWithoutStory(t *testing.T) {
	md, err := storyPage("Before the Map", "").ToMarkdown("")
	assert.FailsNow(t).OK(md, err)
	assert.Continues(t).
		False(strings.Contains(md, "continues")).
		Equal(strings.Count(md, "{{</featured/scrolly-map>}}"), 1)
}

func TestScrollyMapStopHasContent(t *testing.T) {
	assert.Continues(t).
		False(ScrollyMapStop{}.hasContent()).
		False(ScrollyMapStop{Color: "Blue", CardPosition: "Left"}.hasContent()).
		True(ScrollyMapStop{Dots: "Altoona: 40.5, -78.4"}.hasContent()).
		True(ScrollyMapStop{Image: "2026/05/photo.jpeg"}.hasContent()).
		True(ScrollyMapStop{Datawrapper: "1Dg8P"}.hasContent()).
		True(ScrollyMapStop{Flourish: "visualisation/1"}.hasContent()).
		True(ScrollyMapStop{Label: "A"}.hasContent()).
		True(ScrollyMapStop{CardText: "Text"}.hasContent()).
		True(ScrollyMapStop{GeoJSON: "https://example.com/a.geojson"}.hasContent())
}
