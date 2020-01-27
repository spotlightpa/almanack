package feed

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/spotlightpa/almanack/internal/errutil"
	"github.com/spotlightpa/almanack/internal/jsonschema"
)

type Feed struct {
	Stories []*Story
}

func (feed Feed) Get(id string) (*Story, error) {
	found := -1
	for i, article := range feed.Stories {
		if article.ArcID == id {
			if found != -1 {
				return nil, fmt.Errorf("multiple matching IDs found")
			}
			found = i
		}
	}
	if found == -1 {
		return nil, errutil.NotFound
	}
	return feed.Stories[found], nil
}

func (f *Feed) UnmarshalJSON(data []byte) error {
	var v jsonschema.API
	err := json.Unmarshal(data, &v)
	f.Stories = make([]*Story, 0, len(v.Contents))
	for _, content := range v.Contents {
		story := ContentToStory(content)
		if story != nil {
			f.Stories = append(f.Stories, story)
		}
	}
	sort.Slice(f.Stories, func(i, j int) bool {
		return f.Stories[i].PubDate.Before(f.Stories[j].PubDate)
	})
	return err
}

type Story struct {
	ArcID        string    `toml:"arc-id"`
	ID           string    `toml:"internal-id"`
	ImageCredit  string    `toml:"image-credit"`
	ImageCaption string    `toml:"image-description"`
	ImageURL     string    `toml:"image"`
	Slug         string    `toml:"slug"`
	PubDate      time.Time `toml:"published"`
	Budget       string    `toml:"internal-budget"`
	Hed          string    `toml:"title"`
	Subhead      string    `toml:"subtitle"`
	Summary      string    `toml:"description"`
	Blurb        string    `toml:"blurb"`
	Authors      []string  `toml:"authors"`
	Body         string    `toml:"-"`
	LinkTitle    string    `toml:"linktitle"`
}

func ContentToStory(content jsonschema.Contents) *Story {
	authors := make([]string, len(content.Credits.By))
	for i := range content.Credits.By {
		authors[i] = content.Credits.By[i].Name
	}
	var body strings.Builder
	readContentElements(content.ContentElements, &body)
	story := Story{
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
	if image := imageFrom(content.PromoItems); image != nil {
		story.ImageURL = image.URL
		story.ImageCaption = image.Caption
		story.ImageCredit = image.Credit
	}
	return &story
}

func (story *Story) String() string {
	if story == nil {
		return "<nil story>"
	}
	return fmt.Sprintf("%#v", *story)
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

func readContentElements(rawels []*json.RawMessage, body *strings.Builder) {
	for i, raw := range rawels {
		var _type string
		wrapper := jsonschema.ContentElementType{Type: &_type}
		if err := json.Unmarshal(*raw, &wrapper); err != nil {
			log.Printf("runtime error: %v", err)
		}
		var graf string
		switch _type {
		case "text", "raw_html":
			wrapper := jsonschema.ContentElementText{Content: &graf}
			if err := json.Unmarshal(*raw, &wrapper); err != nil {
				log.Printf("runtime error: %v", err)
			}

		case "header":
			var v jsonschema.ContentElementHeading
			if err := json.Unmarshal(*raw, &v); err != nil {
				log.Printf("runtime error: %v", err)
			}
			graf = strings.Repeat("#", v.Level) + " " + v.Content
		case "oembed_response":
			var v jsonschema.ContentElementOembed
			if err := json.Unmarshal(*raw, &v); err != nil {
				log.Printf("runtime error: %v", err)
			}
			graf = v.RawOembed.HTML
		case "list":
			var v jsonschema.ContentElementList
			if err := json.Unmarshal(*raw, &v); err != nil {
				log.Printf("runtime error: %v", err)
			}

			var identifier string
			switch v.ListType {
			case "unordered":
				identifier = "- "
			default:
				log.Printf("warning: unknown list type - %q", v.ListType)
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
					log.Printf("warning: unknown item type - %q", item.Type)
				}
				body.WriteString(identifier)
				body.WriteString(li)
				body.WriteString("\n\n")
			}

		case "image":
			var v jsonschema.ContentElementImage
			if err := json.Unmarshal(*raw, &v); err != nil {
				log.Printf("runtime error: %v", err)
			}
			var credits []string
			for _, c := range v.Credits.By {
				credits = append(credits, c.Name)
			}
			graf = fmt.Sprintf("## Image:\n\n%s\n\n%s (%s)\n",
				v.URL, v.Caption, strings.Join(credits, " "),
			)

		default:
			log.Printf("warning: unknown element type - %q", _type)
		}
		if i != 0 {
			body.WriteString("\n\n")
		}
		body.WriteString(graf)
	}
}

type Image struct {
	Credit, Caption, URL string
}

func imageFrom(p jsonschema.PromoItems) *Image {
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
	return &Image{
		strings.Join(credits, " / "),
		p.Basic.Caption,
		p.Basic.URL,
	}
}

func (story Story) ToTOML() (string, error) {
	var buf strings.Builder
	buf.WriteString("+++\n")
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(story); err != nil {
		return "", err
	}
	buf.WriteString("+++\n\n")
	buf.WriteString(story.Body)
	buf.WriteString("\n")
	return buf.String(), nil
}
