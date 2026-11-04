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
		Slug:             "regions",
		Headline:         "Defining Pennsylvania's Regions",
		Byline:           "Asha Prihar",
		CreditType:       "Words by",
		HeadlineSize:     "Large",
		PublishedAt:      time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC),
		TopperLayout:     "Icon",
		TopperImage:      "2026/09/topper.png",
		TopperBackground: "#f4f1ee",
		CardPosition:     "Right",
		UnshadedColor:    "#e5e5e5",
		BasemapHidden:    true,
		Background:       "Navy",
		OutlineOff:       true,
		Stops: []ScrollyMapStop{
			{Label: "Central PA", GeoJSON: "https://example.com/a.geojson", CardText: "Hello", ZoomTo: "Shaded", ShadingOff: true, ImageRatio: "16:9"},
		},
	}
	md := be.OK(page.ToMarkdown(""))
	for _, want := range []string{
		`credit-type="Words by"`,
		`hed-size="large"`,
		`topper-layout="icon"`,
		`topper-bg="#f4f1ee"`,
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

func TestScrollyMapTopperSkippedWithoutImage(t *testing.T) {
	be := assert.FailsNow(t)
	page := ScrollyMapPage{
		Slug:             "regions",
		Headline:         "Regions",
		TopperLayout:     "Icon",
		TopperBackground: "#f4f1ee",
		Stops:            []ScrollyMapStop{{Label: "A"}},
	}
	md := be.OK(page.ToMarkdown(""))
	be.False(strings.Contains(md, "topper-layout")).
		False(strings.Contains(md, "topper-bg"))
}
