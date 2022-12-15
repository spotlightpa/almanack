package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/pkg/common"
)

func ArcFeedItemToPage(ctx context.Context, svc Services, arcStory *arc.FeedItem, article *SpotlightPAArticle) (err error) {
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
		if author, _, ok := strings.Cut(article.Authors[i], " of "); ok {
			article.Authors[i] = author
		} else if author, _, ok = strings.Cut(article.Authors[i], " for "); ok {
			article.Authors[i] = author
		}
	}

	// Drop "Spotlight PA Staff" as an author
	{
		filteredAuthors := article.Authors[:0]
		for _, author := range article.Authors {
			switch {
			case strings.EqualFold(author, "Spotlight PA Staff"):
			case strings.EqualFold(author, "Spotlight PA State College Staff"):
			default:
				filteredAuthors = append(filteredAuthors, author)
			}
		}
		article.Authors = filteredAuthors
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
func authorFrom(by *arc.By) string {
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
	slug, _, ok := stringx.LastCut(s, "-")
	if !ok {
		return s
	}
	_, slug, ok = stringx.LastCut(slug, "/")
	if !ok {
		return s
	}
	return slug
}

func readContentElements(ctx context.Context, svc Services, rawels []*json.RawMessage, body *strings.Builder) (warnings []string, err error) {
	for i, raw := range rawels {
		var _type string
		wrapper := arc.ContentElementType{Type: &_type}
		if err := json.Unmarshal(*raw, &wrapper); err != nil {
			common.Logger.Printf("runtime error: %v", err)
		}
		var graf string
		switch _type {
		case "text", "raw_html":
			wrapper := arc.ContentElementText{Content: &graf}
			if err := json.Unmarshal(*raw, &wrapper); err != nil {
				return nil, err
			}

		case "header":
			var v arc.ContentElementHeading
			if err := json.Unmarshal(*raw, &v); err != nil {
				common.Logger.Printf("runtime error: %v", err)
			}
			graf = strings.Repeat("#", v.Level) + " " + v.Content
		case "oembed_response":
			var v arc.ContentElementOembed
			if err := json.Unmarshal(*raw, &v); err != nil {
				return nil, err
			}
			graf = v.RawOembed.HTML
		case "list":
			var v arc.ContentElementList
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
			var v arc.ContentElementImage
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

var inkyURL = must.Get(url.Parse("https://www.inquirer.com"))

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

func setArticleImage(a *SpotlightPAArticle, p arc.PromoItems) {
	a.ImageURL = resolveFromInky(p.Basic.AdditionalProperties.ResizeURL)
	if a.ImageURL == "" && strings.Contains(p.Basic.URL, "public") {
		a.ImageURL = resolveFromInky(p.Basic.URL)
	}
	var credits []string
	for _, credit := range p.Basic.Credits.By {
		credits = append(credits, stringx.First(credit.Name, credit.Byline))
	}
	a.ImageCredit = fixCredit(strings.Join(credits, " / "))
	a.ImageDescription = p.Basic.Caption
}

var fixcreditre = regexp.MustCompile(`(?i)\b(staff( photographer)?)\b`)

// change staff to inky
func fixCredit(s string) string {
	return fixcreditre.ReplaceAllLiteralString(s, "Philadelphia Inquirer")
}
