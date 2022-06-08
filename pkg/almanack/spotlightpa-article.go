package almanack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spotlightpa/almanack/internal/db"
)

type SpotlightPAArticle struct {
	ArcID            string     `toml:"arc-id"`
	InternalID       string     `toml:"internal-id"`
	Budget           string     `toml:"internal-budget"`
	ImageURL         string     `toml:"image"`
	ImageDescription string     `toml:"image-description"`
	ImageCaption     string     `toml:"image-caption"`
	ImageCredit      string     `toml:"image-credit"`
	ImageSize        string     `toml:"image-size"`
	PubDate          time.Time  `toml:"published"`
	Slug             string     `toml:"slug"`
	Authors          []string   `toml:"authors"`
	Byline           string     `toml:"byline"`
	Hed              string     `toml:"title"`
	Subhead          string     `toml:"subtitle"`
	Summary          string     `toml:"description"`
	Blurb            string     `toml:"blurb"`
	Kicker           string     `toml:"kicker"`
	Topics           []string   `toml:"topics"`
	Series           []string   `toml:"series"`
	LinkTitle        string     `toml:"linktitle"`
	SuppressFeatured bool       `toml:"-"`
	Weight           int        `toml:"-"`
	OverrideURL      string     `toml:"url"`
	Aliases          []string   `toml:"aliases"`
	ModalExclude     bool       `toml:"modal-exclude"`
	NoIndex          bool       `toml:"no-index"`
	LanguageCode     string     `toml:"language-code"`
	LayoutType       string     `toml:"layout"`
	ExtendedKicker   string     `toml:"extended-kicker"`
	Body             string     `toml:"-"`
	Filepath         string     `toml:"-"`
	LastArcSync      time.Time  `toml:"-"`
	ScheduleFor      *time.Time `toml:"-"`
	LastSaved        *time.Time `toml:"-"`
	LastPublished    *time.Time `toml:"-"`
	Warnings         []string   `toml:"-"`
	PageKind         string     `toml:"-"`
}

func (splArt *SpotlightPAArticle) ResetArcData(ctx context.Context, svc Service, dbArticle db.Article) (err error) {
	var arcStory ArcStory
	if err = arcStory.fromDB(&dbArticle); err != nil {
		return err
	}

	if err = arcStory.ToArticle(ctx, svc, splArt); err == nil {
		splArt.LastArcSync = dbArticle.UpdatedAt
	}
	return nil
}

func (splArt *SpotlightPAArticle) String() string {
	if splArt == nil {
		return "<nil article>"
	}
	return fmt.Sprintf("%#v", *splArt)
}

func (splArt *SpotlightPAArticle) URL() string {
	if splArt.Slug == "" || splArt.PubDate.IsZero() {
		return ""
	}
	pagekind := "news"
	if splArt.PageKind != "" {
		pagekind = splArt.PageKind
	}
	year := splArt.PubDate.Year()
	month := splArt.PubDate.Month()
	return fmt.Sprintf(
		"https://www.spotlightpa.org/%s/%d/%02d/%s/",
		pagekind, year, month, splArt.Slug,
	)
}

func (splArt *SpotlightPAArticle) ContentFilepath() string {
	pagekind := "news"
	if splArt.PageKind != "" {
		pagekind = splArt.PageKind
	}
	if splArt.Filepath == "" {
		date := splArt.PubDate.Format("2006-01-02")
		splArt.Filepath = fmt.Sprintf("content/%s/%s-%s.md",
			pagekind, date, splArt.InternalID)
	}
	return splArt.Filepath
}

func (splArt *SpotlightPAArticle) ToTOML() (string, error) {
	var buf strings.Builder
	buf.WriteString("+++\n")
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(splArt); err != nil {
		return "", err
	}
	buf.WriteString("+++\n\n")
	buf.WriteString(splArt.Body)
	buf.WriteString("\n")
	return buf.String(), nil
}

func (splArt *SpotlightPAArticle) FromTOML(content string) error {
	const delimiter = "+++\n"

	if !strings.HasPrefix(content, delimiter) {
		return fmt.Errorf("could not parse frontmatter: no prefix delimiter")
	}
	content = strings.TrimPrefix(content, delimiter)
	frontmatter, body, ok := strings.Cut(content, delimiter)
	if !ok {
		return fmt.Errorf("could not parse frontmatter: no end delimiter")
	}

	if _, err := toml.Decode(frontmatter, splArt); err != nil {
		return err
	}
	body = strings.TrimPrefix(body, "\n")
	body = strings.TrimSuffix(body, "\n")
	splArt.Body = body
	return nil
}
