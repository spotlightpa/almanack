package anf

import (
	"cmp"

	"github.com/spotlightpa/almanack/internal/db"
)

func FromDB(item *db.NewsFeedItem) (*Article, error) {
	a, err := ConvertToAppleNews(item.ContentHtml, item.URL)
	if err != nil {
		return nil, err
	}
	comps := a.Components
	a.Identifier = item.ExternalID
	a.Title = item.Title
	a.Language = cmp.Or(item.Language, "en-us")
	a.Metadata = &Metadata{
		Authors:      item.Authors,
		CanonicalURL: item.URL,
		Excerpt:      item.Blurb,
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
			Role: "byline",
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
	return a, nil
}

func BuildSections(item *db.NewsFeedItem, sectionToURL map[string]string) []string {
	var sections []string
	// Add default section
	sections = append(sections, sectionToURL[""])
	for _, topic := range item.Topics {
		if u, ok := sectionToURL[topic]; ok {
			sections = append(sections, u)
		}
	}
	return sections
}
