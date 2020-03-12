package almanack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Article struct {
	ArcID            string    `toml:"arc-id"`
	InternalID       string    `toml:"internal-id"`
	Budget           string    `toml:"internal-budget"`
	ImageURL         string    `toml:"image"`
	ImageCaption     string    `toml:"image-description"`
	ImageCredit      string    `toml:"image-credit"`
	PubDate          time.Time `toml:"published"`
	Slug             string    `toml:"slug"`
	Authors          []string  `toml:"authors"`
	Byline           string    `toml:"byline"`
	Hed              string    `toml:"title"`
	Subhead          string    `toml:"subtitle"`
	Summary          string    `toml:"description"`
	Blurb            string    `toml:"blurb"`
	Kicker           string    `toml:"kicker"`
	Body             string    `toml:"-"`
	LinkTitle        string    `toml:"linktitle"`
	SuppressFeatured bool      `toml:"suppress-featured"`
}

func (article *Article) String() string {
	if article == nil {
		return "<nil article>"
	}
	return fmt.Sprintf("%#v", *article)
}

func (article *Article) ContentFilepath() string {
	date := article.PubDate.Format("2006-01-02")
	return fmt.Sprintf("content/news/%s-%s.md", date, article.InternalID)
}

func (article *Article) ToTOML() (string, error) {
	var buf strings.Builder
	buf.WriteString("+++\n")
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(article); err != nil {
		return "", err
	}
	buf.WriteString("+++\n\n")
	buf.WriteString(article.Body)
	buf.WriteString("\n")
	return buf.String(), nil
}

func (article *Article) Publish(ctx context.Context, gh ContentStore) error {
	data, err := article.ToTOML()
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Content: publishing %q", article.InternalID)
	path := article.ContentFilepath()
	if err = gh.CreateFile(ctx, msg, path, []byte(data)); err != nil {
		return err
	}
	return nil
}
