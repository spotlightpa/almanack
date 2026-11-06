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
		UnshadedColor: "#e5e5e5",
		BasemapHidden: true,
		Background:    "Navy",
		OutlineOff:    true,
		Stops: []ScrollyMapStop{
			{Label: "Central PA", GeoJSON: "https://example.com/a.geojson", CardText: "Hello", ZoomTo: "Shaded", ShadingOff: true, ImageRatio: "16:9"},
		},
	}
	md := be.OK(page.ToMarkdown(""))
	for _, want := range []string{
		`image-gravity = "ce"`,
		`topper-layout="icon-light"`,
		`base-color="#e5e5e5"`,
		`layout="right"`,
		`basemap="hide"`,
		`background="Navy"`,
		`outline="false"`,
		`fit="shaded"`,
		`shade="false"`,
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
