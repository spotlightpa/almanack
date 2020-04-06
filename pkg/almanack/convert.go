package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"
)

func (arcStory *ArcStory) ToArticle(ctx context.Context, svc Service, article *SpotlightPAArticle) error {
	article.Authors = make([]string, len(arcStory.Credits.By))
	needsByline := false
	for i := range arcStory.Credits.By {
		article.Authors[i] = authorFrom(&arcStory.Credits.By[i])
		if strings.HasSuffix(article.Authors[i], " of Spotlight PA") {
			needsByline = true
		}
	}
	if needsByline {
		article.Byline = commaAndJoiner(article.Authors)
		for i := range article.Authors {
			article.Authors[i] = strings.TrimSuffix(article.Authors[i], " of Spotlight PA")
		}
	}

	var body strings.Builder
	if err := readContentElements(ctx, svc, arcStory.ContentElements, &body); err != nil {
		return err
	}

	article.ArcID = arcStory.ID
	article.InternalID = arcStory.Slug
	// Don't reset slug on saved stories
	if strings.TrimSpace(article.Slug) == "" {
		article.Slug = slugFromURL(arcStory.CanonicalURL)
	}
	article.PubDate = arcStory.Planning.Scheduling.PlannedPublishDate
	article.Budget = arcStory.Planning.BudgetLine
	article.Hed = arcStory.Headlines.Basic
	article.Subhead = arcStory.Subheadlines.Basic
	article.Summary = arcStory.Description.Basic
	article.Blurb = arcStory.Description.Basic
	article.Body = body.String()
	article.LinkTitle = arcStory.Headlines.Web

	setArticleImage(article, arcStory.PromoItems)
	return nil
}

// Must keep in sync with Vue's ArcArticle.authors
func authorFrom(by *By) string {
	byline := by.AdditionalProperties.Original.Byline
	if byline != "" {
		return byline
	}
	byline = by.Name
	// Hack for bad names with orgs in them
	if strings.Contains(byline, " of ") {
		return byline
	}
	if org := strings.TrimSpace(by.Org); org != "" {
		return byline + " of " + org
	}
	return byline
}

func commaAndJoiner(ss []string) string {
	if len(ss) < 3 {
		return strings.Join(ss, " and ")
	}
	commaPart := strings.Join(ss[:len(ss)-1], ", ")
	return commaPart + " and " + ss[len(ss)-1]
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

func readContentElements(ctx context.Context, svc Service, rawels []*json.RawMessage, body *strings.Builder) error {
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

			credit := strings.Join(credits, " ")

			u, err := svc.ReplaceImageURL(ctx, v.URL, v.Caption, credit)
			if err != nil {
				return err
			}
			u = html.EscapeString(u)
			desc := html.EscapeString(v.Caption)
			credit = html.EscapeString(credit)
			graf = fmt.Sprintf(
				`{{<picture src="%s" description="%s" caption="%s" credit="%s">}}`+"\n",
				u, desc, desc, credit,
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

func setArticleImage(a *SpotlightPAArticle, p PromoItems) {
	var credits []string
	if strings.Contains(p.Basic.URL, "public") {
		a.ImageURL = p.Basic.URL
	} else {
		a.ImageURL = p.Basic.AdditionalProperties.ResizeURL
	}
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
	re := regexp.MustCompile(`(?i)\b(staff( photographer)?)\b`)
	a.ImageCredit = re.ReplaceAllLiteralString(a.ImageCredit, "Philadelphia Inquirer")
	a.ImageDescription = p.Basic.Caption
}
