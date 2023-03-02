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
	"github.com/spotlightpa/almanack/internal/slicex"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func ArcFeedItemToBody(ctx context.Context, svc Services, arcStory *arc.FeedItem) (body string, warnings []string, err error) {
	var buf strings.Builder
	if warnings, err = readContentElements(ctx, svc, arcStory.ContentElements, &buf); err != nil {
		return
	}
	body = buf.String()
	return
}

func ArcFeedItemToFrontmatter(ctx context.Context, svc Services, arcStory *arc.FeedItem) (fm map[string]any, err error) {
	fm = make(map[string]any)

	// Hacky: Add the of/for XX orgs then remove them
	authors := make([]string, len(arcStory.Credits.By))
	for i := range arcStory.Credits.By {
		authors[i] = authorFrom(&arcStory.Credits.By[i])
	}

	fm["byline"] = commaAndJoiner(authors)

	for i := range authors {
		if author, _, ok := strings.Cut(authors[i], " of "); ok {
			authors[i] = author
		} else if author, _, ok = strings.Cut(authors[i], " for "); ok {
			authors[i] = author
		}
	}

	// Drop "Spotlight PA Staff" as an author
	slicex.DeleteFunc(&authors, func(author string) bool {
		return strings.EqualFold(author, "Spotlight PA Staff") ||
			strings.EqualFold(author, "Spotlight PA State College Staff")
	})

	fm["authors"] = authors

	fm["arc-id"] = arcStory.ID
	fm["internal-id"] = arcStory.Slug

	fm["slug"] = slugFromURL(arcStory.CanonicalURL)
	fm["published"] = arcStory.Planning.Scheduling.PlannedPublishDate
	fm["internal-budget"] = arcStory.Planning.BudgetLine
	fm["title"] = arcStory.Headlines.Basic
	// Subtitle isn't exposed in the current layout
	// fm["subtitle"] = arcStory.Subheadlines.Basic
	fm["description"] = arcStory.Description.Basic
	fm["blurb"] = arcStory.Description.Basic
	fm["linktitle"] = arcStory.Headlines.Web

	p := arcStory.PromoItems
	imageURL := resolveFromInky(p.Basic.AdditionalProperties.ResizeURL)
	if imageURL == "" && strings.Contains(p.Basic.URL, "public") {
		imageURL = resolveFromInky(p.Basic.URL)
	}
	var credits []string
	for _, credit := range p.Basic.Credits.By {
		credits = append(credits, stringx.First(credit.Name, credit.Byline))
	}
	imageCredit := fixCredit(strings.Join(credits, " / "))
	imageDescription := p.Basic.Caption

	if imageURL != "" {
		var imgerr error
		imageURL, imgerr = svc.ReplaceImageURL(
			ctx, imageURL, imageDescription, imageCredit)
		if imgerr != nil {
			return nil, imgerr
		}
	}

	fm["image"] = imageURL
	fm["image-credit"] = imageCredit
	fm["image-description"] = imageDescription

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
	l := almlog.FromContext(ctx)
	for i, raw := range rawels {
		var _type string
		wrapper := arc.ContentElementType{Type: &_type}
		if err := json.Unmarshal(*raw, &wrapper); err != nil {
			l.ErrorCtx(ctx, "readContentElements unwrap", err)
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
				l.ErrorCtx(ctx, "readContentElements header", err)
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
				return nil, imgerr
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

var fixcreditre = regexp.MustCompile(`(?i)\b(staff( photographer)?)\b`)

// change staff to inky
func fixCredit(s string) string {
	return fixcreditre.ReplaceAllLiteralString(s, "Philadelphia Inquirer")
}
