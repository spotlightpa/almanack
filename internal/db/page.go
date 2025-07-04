package db

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/carlmjohnson/errorx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/microcosm-cc/bluemonday"
	"github.com/spotlightpa/almanack/internal/timex"
)

func CreatePageFromContent(ctx context.Context, tx *Txable, path, content string) (*Page, error) {
	var p Page
	if err := p.FromTOML(content); err != nil {
		return nil, err
	}
	err := tx.Begin(ctx, pgx.TxOptions{}, func(q *Queries) error {
		dbPage, err := q.CreatePage(ctx, CreatePageParams{
			FilePath:   path,
			SourceType: "load",
			SourceID:   path,
		})
		if err != nil {
			return err
		}

		// Fix p fields
		p.SetURLPath()

		pubDate, ok := timex.Unwrap(p.Frontmatter["published"])
		if !ok {
			pubDate = time.Now()
		}
		pubDate = timex.ToEST(pubDate)
		p.Frontmatter["published"] = pubDate

		dbPage, err = q.UpdatePage(ctx, UpdatePageParams{
			ID:               dbPage.ID,
			SetFrontmatter:   true,
			Frontmatter:      p.Frontmatter,
			SetBody:          true,
			Body:             p.Body,
			URLPath:          p.URLPath.String,
			SetScheduleFor:   true,
			ScheduleFor:      pgtype.Timestamptz{Time: pubDate, Valid: true},
			SetLastPublished: true,
		})
		p = dbPage
		return err
	})
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (page *Page) ToTOML() (string, error) {
	var buf strings.Builder
	buf.WriteString("+++\n")
	enc := toml.NewEncoder(&buf)
	// Remove blank values
	frontmatter := Map{}
	for key, val := range page.Frontmatter {
		if val == nil {
			continue
		}
		if s, ok := val.(string); ok && s == "" {
			continue
		}
		if t, ok := val.(time.Time); ok && t.IsZero() {
			continue
		}
		if n, ok := val.(float64); ok && n == 0.0 {
			continue
		}
		if n, ok := val.(int64); ok && n == 0 {
			continue
		}
		if v := reflect.ValueOf(val); v.Kind() == reflect.Slice &&
			v.Len() == 0 {
			continue
		}
		if t, ok := timex.Unwrap(page.Frontmatter[key]); ok {
			val = timex.ToEST(t)
		}
		frontmatter[key] = val
	}
	if err := enc.Encode(frontmatter); err != nil {
		return "", err
	}
	buf.WriteString("+++\n\n")
	buf.WriteString(page.Body)
	buf.WriteString("\n")
	return buf.String(), nil
}

func (page *Page) FromTOML(content string) (err error) {
	defer errorx.Trace(&err)

	const delimiter = "+++\n"
	content, ok := strings.CutPrefix(content, delimiter)
	if !ok {
		// try parsing as JSON
		if !strings.HasPrefix(content, "{") {
			return fmt.Errorf("could not parse frontmatter: no prefix delimiter")
		}
		m := map[string]any{}
		if err := json.Unmarshal([]byte(content), &m); err != nil {
			return err
		}
		page.Frontmatter = m
		page.Body = ""
		return nil
	}
	frontmatter, body, ok := strings.Cut(content, delimiter)
	if !ok {
		return fmt.Errorf("could not parse frontmatter: no end delimiter")
	}

	m := map[string]any{}
	if _, err := toml.Decode(frontmatter, &m); err != nil {
		return err
	}
	page.Frontmatter = m
	body = strings.TrimPrefix(body, "\n")
	body = strings.TrimSuffix(body, "\n")
	page.Body = body
	return nil
}

func (page *Page) SetURLPath() {
	if page.URLPath.Valid && page.URLPath.String != "" {
		return
	}
	if u, _ := page.Frontmatter["url"].(string); u != "" {
		page.URLPath.String = strings.ToLower(u)
		page.URLPath.Valid = true
		return
	}
	upath := page.FilePath
	upath = strings.TrimPrefix(upath, "content")
	upath = strings.TrimSuffix(upath, ".md")
	dir, fname := path.Split(upath)
	if dir == "/news/" || dir == "/statecollege/" || dir == "/berks/" {
		if pub, ok := timex.Unwrap(page.Frontmatter["published"]); ok {
			pub = timex.ToEST(pub)
			dir = pub.Format(dir + "2006/01/")
		}
	}
	if slug, _ := page.Frontmatter["slug"].(string); slug != "" {
		fname = slug
	}

	upath = path.Join(dir, fname)
	if upath != "" && !strings.HasSuffix(upath, "/") {
		upath += "/"
	}
	page.URLPath.String = strings.ToLower(upath)
	page.URLPath.Valid = upath != ""
}

func (page *Page) FullURL() string {
	page.SetURLPath()
	if !page.URLPath.Valid {
		return ""
	}
	return fmt.Sprintf("https://www.spotlightpa.org%s", page.URLPath.String)
}

func (page *Page) ToIndex() any {
	internalID, _ := page.Frontmatter["internal-id"].(string)
	imageURL, _ := page.Frontmatter["image"].(string)
	imageDescription, _ := page.Frontmatter["image-description"].(string)
	imageCaption, _ := page.Frontmatter["image-caption"].(string)
	imageCredit, _ := page.Frontmatter["image-credit"].(string)
	imageSize, _ := page.Frontmatter["image-size"].(string)
	pubDate, _ := timex.Unwrap(page.Frontmatter["published"])
	slug, _ := page.Frontmatter["slug"].(string)
	authors, _ := page.Frontmatter["authors"].([]string)
	byline, _ := page.Frontmatter["byline"].(string)
	hed, _ := page.Frontmatter["title"].(string)
	// subhead is unused?
	subhead, _ := page.Frontmatter["description"].(string)
	summary, _ := page.Frontmatter["summary"].(string)
	blurb, _ := page.Frontmatter["blurb"].(string)
	kicker, _ := page.Frontmatter["kicker"].(string)
	topics, _ := page.Frontmatter["topics"].([]string)
	series, _ := page.Frontmatter["series"].([]string)
	linkTitle, _ := page.Frontmatter["linkTitle"].(string)
	aliases, _ := page.Frontmatter["aliases"].([]string)
	rawContent, _ := page.Frontmatter["raw-content"].(string)

	body := cmp.Or(page.Body, rawContent)
	// Strip any unorthodox HTML
	sanitizer := bluemonday.UGCPolicy()
	body = sanitizer.Sanitize(body)
	// See https://www.algolia.com/doc/guides/sending-and-managing-data/prepare-your-data/in-depth/index-and-records-size-and-usage-limitations/#record-size-limits
	const maxLen = 80_000
	if len(body) > maxLen {
		body = body[:maxLen]
	}
	return struct {
		ObjectID         string    `json:"objectID"`
		URL              string    `json:"URL"`
		InternalID       string    `json:"internal-id"`
		ImageURL         string    `json:"image-url"`
		ImageDescription string    `json:"image-description"`
		ImageCaption     string    `json:"image-caption"`
		ImageCredit      string    `json:"image-credit"`
		ImageSize        string    `json:"image-size"`
		PubDate          time.Time `json:"pub-date"`
		Slug             string    `json:"slug"`
		Authors          []string  `json:"authors"`
		Byline           string    `json:"byline"`
		Hed              string    `json:"hed"`
		Subhead          string    `json:"subhead"`
		Summary          string    `json:"summary"`
		Blurb            string    `json:"blurb"`
		Kicker           string    `json:"kicker"`
		Topics           []string  `json:"topics"`
		Series           []string  `json:"series"`
		LinkTitle        string    `json:"link-title"`
		Aliases          []string  `json:"aliases"`
		Body             string    `json:"body"`
	}{
		page.FullURL(),
		page.FullURL(),
		internalID,
		imageURL,
		imageDescription,
		imageCaption,
		imageCredit,
		imageSize,
		pubDate,
		slug,
		authors,
		byline,
		hed,
		subhead,
		summary,
		blurb,
		kicker,
		topics,
		series,
		linkTitle,
		aliases,
		body,
	}
}

func (page *Page) ShouldPublish() bool {
	soon := time.Now().Add(5 * time.Minute)
	isScheduled := page.ScheduleFor.Valid
	return isScheduled && page.ScheduleFor.Time.Before(soon)
}

func (page *Page) IsNewsyPage() bool {
	return page.IsNewsPage() || page.IsStateCollegePage() || page.IsBerksPage()
}

func (page *Page) IsNewsPage() bool {
	return strings.HasPrefix(page.FilePath, "content/news/")
}

func (page *Page) IsStateCollegePage() bool {
	return strings.HasPrefix(page.FilePath, "content/statecollege/")
}

func (page *Page) IsBerksPage() bool {
	return strings.HasPrefix(page.FilePath, "content/berks/")
}

func (page *Page) ShouldNotify(oldPage *Page) bool {
	if !page.IsNewsyPage() || !page.ScheduleFor.Valid {
		return false
	}
	if page.ShouldPublish() {
		return !oldPage.LastPublished.Valid
	}

	return !timex.Equalish(oldPage.ScheduleFor, page.ScheduleFor)
}
