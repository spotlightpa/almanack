package arcjson

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/spotlightpa/almanack/pkg/almanack"
)

func (content Contents) ToArticle() (*almanack.Article, error) {
	authors := make([]string, len(content.Credits.By))
	for i := range content.Credits.By {
		authors[i] = content.Credits.By[i].Name
	}
	var body strings.Builder
	if err := readContentElements(content.ContentElements, &body); err != nil {
		return nil, err
	}
	story := almanack.Article{
		ArcID:     content.ID,
		ID:        content.Slug,
		Slug:      slugFromURL(content.CanonicalURL),
		PubDate:   content.Planning.Scheduling.PlannedPublishDate,
		Budget:    content.Planning.BudgetLine,
		Hed:       content.Headlines.Basic,
		Subhead:   content.Subheadlines.Basic,
		Summary:   content.Description.Basic,
		Blurb:     content.Description.Basic,
		Authors:   authors,
		Body:      body.String(),
		LinkTitle: content.Headlines.Web,
	}
	imageFrom(&story, content.PromoItems)
	return &story, nil
}

func slugFromURL(s string) string {
	stop := strings.LastIndexByte(s, '-')
	if stop == -1 {
		return s
	}
	start := strings.LastIndexByte(s[:stop], '/')
	if start == -1 {
		return s
	}
	return s[start+1 : stop]
}

func readContentElements(rawels []*json.RawMessage, body *strings.Builder) error {
	for i, raw := range rawels {
		var _type string
		wrapper := ContentElementType{Type: &_type}
		if err := json.Unmarshal(*raw, &wrapper); err != nil {
			log.Printf("runtime error: %v", err)
		}
		var graf string
		switch _type {
		case "text", "raw_html":
			wrapper := ContentElementText{Content: &graf}
			if err := json.Unmarshal(*raw, &wrapper); err != nil {
				return err
			}

		case "header":
			var v ContentElementHeading
			if err := json.Unmarshal(*raw, &v); err != nil {
				log.Printf("runtime error: %v", err)
			}
			graf = strings.Repeat("#", v.Level) + " " + v.Content
		case "oembed_response":
			var v ContentElementOembed
			if err := json.Unmarshal(*raw, &v); err != nil {
				return err
			}
			graf = v.RawOembed.HTML
		case "list":
			var v ContentElementList
			if err := json.Unmarshal(*raw, &v); err != nil {
				return err
			}

			var identifier string
			switch v.ListType {
			case "unordered":
				identifier = "- "
			default:
				return fmt.Errorf("unkown list type: %q", v.ListType)
			}
			for j, item := range v.Items {
				var li string
				if j != 0 {
					body.WriteString("\n\n")
				}
				switch item.Type {
				case "text":
					li = item.Content
				default:
					return fmt.Errorf("unknown item type: %q", item.Type)
				}
				body.WriteString(identifier)
				body.WriteString(li)
				body.WriteString("\n\n")
			}

		case "image":
			var v ContentElementImage
			if err := json.Unmarshal(*raw, &v); err != nil {
				return err
			}
			var credits []string
			for _, c := range v.Credits.By {
				credits = append(credits, c.Name)
			}
			graf = fmt.Sprintf("## Image:\n\n%s\n\n%s (%s)\n",
				v.URL, v.Caption, strings.Join(credits, " "),
			)

		default:
			return fmt.Errorf("unknown element type - %q", _type)
		}
		if i != 0 {
			body.WriteString("\n\n")
		}
		body.WriteString(graf)
	}
	return nil
}

func imageFrom(a *almanack.Article, p PromoItems) {
	var credits []string
	for i, credit := range p.Basic.Credits.By {
		name := credit.Byline
		if name == "" {
			name = credit.Name
		}
		credits = append(credits, name)
		if len(p.Basic.Credits.Affiliation) > i {
			credits = append(credits, p.Basic.Credits.Affiliation[i].Name)
		}
	}
	a.ImageCredit = strings.Join(credits, " / ")
	a.ImageCaption = p.Basic.Caption
	a.ImageURL = p.Basic.URL
}
