package anf

import (
	"cmp"
	"fmt"
	"strings"

	"github.com/spotlightpa/almanack/internal/db"
	"golang.org/x/net/html"
)

func FromDB(item db.AppleNewsFeed) (*Article, error) {
	comps, err := buildComponents(item.ContentHtml)
	if err != nil {
		return nil, err
	}
	a := templateDoc.Clone()
	a.Identifier = item.ExternalID
	a.Title = item.Title
	a.Language = cmp.Or(item.Language, "en-us")
	a.Metadata = &Metadata{
		Authors:      item.Authors,
		CanonicalURL: item.URL,
		// Excerpt:             "",
		Keywords:     []string{item.Category},
		ThumbnailURL: item.Image,
		// TransparentToolbar:  "",
		// VideoURL:            "",
		// Links:               "",
		DateCreated:         &item.ExternalPublishedAt,
		DateModified:        &item.ExternalUpdatedAt,
		DatePublished:       &item.ExternalPublishedAt,
		GeneratorName:       "Spotlight PA Feed2ANF",
		GeneratorVersion:    "0.0.1",
		GeneratorIdentifier: "",
	}
	cover := FillMode("cover")
	center := VerticalAlignment("center")
	a.Components = []Component{
		TextComponent{
			Role:      "intro",
			Layout:    "eyebrowLayout",
			TextStyle: "eyebrowStyle",
			Text:      "\u00a0" + item.Category + "\u00a0",
		},
		TextComponent{
			Role:   "title",
			Text:   item.Title,
			Layout: "titleLayout",
		},
		TextComponent{
			Role: "author",
			Text: "by " + item.Author,
		},
		TextComponent{
			Role:   "header",
			Layout: "headerImageLayout",
			Style: ComponentStyle{
				Fill: ImageFill{
					Type:              "image",
					URL:               item.Image,
					FillMode:          &cover,
					VerticalAlignment: &center,
				},
			},
		},
	}

	a.Components = append(a.Components, comps...)
	return &a, nil
}

func buildComponents(htmlContent string) ([]Component, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("BuildComponents: parsing HTML: %w", err)
	}
	art := ConvertHTMLToAppleNews(doc)
	return art.Components, nil
}
