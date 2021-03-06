package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/spotlightpa/almanack/internal/stringutils"
)

func (arcStory *ArcStory) ToArticle(ctx context.Context, svc Service, article *SpotlightPAArticle) (err error) {
	var body strings.Builder
	if article.Warnings, err = readContentElements(ctx, svc, arcStory.ContentElements, &body); err != nil {
		return
	}
	article.Body = body.String()
	if len(article.Warnings) > 0 {
		article.ScheduleFor = nil
	}

	// Don't process anything else if this has been saved before
	if !article.LastArcSync.IsZero() {
		return
	}

	// Hacky: Add the of/for XX orgs then remove them
	article.Authors = make([]string, len(arcStory.Credits.By))
	for i := range arcStory.Credits.By {
		article.Authors[i] = authorFrom(&arcStory.Credits.By[i])
	}

	article.Byline = commaAndJoiner(article.Authors)
	for i := range article.Authors {
		if pos := strings.Index(article.Authors[i], " of "); pos != -1 {
			article.Authors[i] = article.Authors[i][:pos]
		} else if pos := strings.Index(article.Authors[i], " for "); pos != -1 {
			article.Authors[i] = article.Authors[i][:pos]
		}
	}

	article.ArcID = arcStory.ID
	article.InternalID = arcStory.Slug

	article.Slug = slugFromURL(arcStory.CanonicalURL)
	article.PubDate = arcStory.Planning.Scheduling.PlannedPublishDate
	article.Budget = arcStory.Planning.BudgetLine
	article.Hed = arcStory.Headlines.Basic
	article.Subhead = arcStory.Subheadlines.Basic
	article.Summary = arcStory.Description.Basic
	article.Blurb = arcStory.Description.Basic
	article.LinkTitle = arcStory.Headlines.Web

	setArticleImage(article, arcStory.PromoItems)
	if strings.HasPrefix(article.ImageURL, "http") {
		var imgerr error
		article.ImageURL, imgerr = svc.ReplaceImageURL(
			ctx, article.ImageURL, article.ImageDescription, article.ImageCredit)
		if imgerr != nil {
			article.Warnings = append(article.Warnings, imgerr.Error())
		}
	}

	if len(article.Warnings) > 0 {
		article.ScheduleFor = nil
	}

	return
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

func readContentElements(ctx context.Context, svc Service, rawels []*json.RawMessage, body *strings.Builder) (warnings []string, err error) {
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
				return nil, err
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
				return nil, err
			}
			graf = v.RawOembed.HTML
		case "list":
			var v ContentElementList
			if err := json.Unmarshal(*raw, &v); err != nil {
				return nil, err
			}

			var buf strings.Builder
			n := 0
			switch v.ListType {
			case "unordered":
				n = -1
			case "ordered":
				n = 1
			default:
				warnings = append(warnings,
					fmt.Sprintf("unknown list type: %q", v.ListType))
				continue
			}
			for j, item := range v.Items {
				if j != 0 {
					buf.WriteString("\n\n")
				}

				var li string
				switch item.Type {
				case "text":
					li = strings.TrimSpace(item.Content)
				default:
					warnings = append(warnings,
						fmt.Sprintf("unknown list type: %q", v.ListType))
					continue
				}
				if n < 1 {
					buf.WriteString("- ")
				} else {
					buf.WriteString(strconv.Itoa(n))
					buf.WriteString(". ")
					n++
				}
				buf.WriteString(li)
			}
			graf = buf.String()

		case "image":
			var v ContentElementImage
			if err := json.Unmarshal(*raw, &v); err != nil {
				return nil, err
			}
			var credits []string
			for _, c := range v.Credits.By {
				credits = append(credits, c.Name)
			}
			credit := fixCredit(strings.Join(credits, " "))

			imageURL := resolveFromInky(v.AdditionalProperties.ResizeURL)
			if imageURL == "" && strings.Contains(v.URL, "public") {
				imageURL = resolveFromInky(v.URL)
			}
			if imageURL == "" {
				warnings = append(warnings,
					fmt.Sprintf("could not find public image for %q", v.URL))
				continue
			}
			u, imgerr := svc.ReplaceImageURL(ctx, imageURL, v.Caption, credit)
			if imgerr != nil {
				warnings = append(warnings, imgerr.Error())
			}
			u = html.EscapeString(u)
			desc := html.EscapeString(v.Caption)
			credit = html.EscapeString(credit)
			graf = fmt.Sprintf(
				`{{<picture src="%s" description="%s" caption="%s" credit="%s">}}`+"\n",
				u, desc, desc, credit,
			)
			graf = strings.ReplaceAll(graf, "\n", " ")

		case "gallery", "interstitial_link":
			continue

		case "divider":
			graf = "<hr>"

		default:
			warnings = append(warnings,
				fmt.Sprintf("unknown element type - %q", _type))
			continue
		}
		if i != 0 {
			body.WriteString("\n\n")
		}
		body.WriteString(graf)
	}
	return
}

var inkyURL = func() *url.URL {
	u, err := url.Parse("https://www.inquirer.com")
	if err != nil {
		panic(err)
	}
	return u
}()

func resolveFromInky(s string) string {
	if s == "" {
		return s
	}
	u, err := inkyURL.Parse(s)
	if err != nil {
		return s
	}
	return u.String()
}

func setArticleImage(a *SpotlightPAArticle, p PromoItems) {
	a.ImageURL = resolveFromInky(p.Basic.AdditionalProperties.ResizeURL)
	if a.ImageURL == "" && strings.Contains(p.Basic.URL, "public") {
		a.ImageURL = resolveFromInky(p.Basic.URL)
	}
	var credits []string
	for _, credit := range p.Basic.Credits.By {
		credits = append(credits, stringutils.First(credit.Name, credit.Byline))
	}
	a.ImageCredit = fixCredit(strings.Join(credits, " / "))
	a.ImageDescription = p.Basic.Caption
}

var fixcreditre = regexp.MustCompile(`(?i)\b(staff( photographer)?)\b`)

// change staff to inky
func fixCredit(s string) string {
	return fixcreditre.ReplaceAllLiteralString(s, "Philadelphia Inquirer")
}
