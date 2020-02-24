package almanack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/spotlightpa/almanack/pkg/errutil"
)

type ArticleService interface {
	GetArticle(id string) (*Article, error)
}

type Feed struct {
	Articles []*Article
}

func (feed Feed) Get(id string) (*Article, error) {
	found := -1
	for i, article := range feed.Articles {
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
	return feed.Articles[found], nil
}

type Article struct {
	ArcID            string    `toml:"arc-id"`
	ID               string    `toml:"internal-id"`
	ImageCredit      string    `toml:"image-credit"`
	ImageCaption     string    `toml:"image-description"`
	ImageURL         string    `toml:"image"`
	Slug             string    `toml:"slug"`
	PubDate          time.Time `toml:"published"`
	Budget           string    `toml:"internal-budget"`
	Hed              string    `toml:"title"`
	Subhead          string    `toml:"subtitle"`
	Summary          string    `toml:"description"`
	Blurb            string    `toml:"blurb"`
	Authors          []string  `toml:"authors"`
	Body             string    `toml:"-"`
	LinkTitle        string    `toml:"linktitle"`
	Byline           string    `toml:"byline"`
	Kicker           string    `toml:"kicker"`
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
	return fmt.Sprintf("content/news/%s-%s.md", date, article.ID)
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
	msg := fmt.Sprintf("Content: publishing %q", article.ID)
	path := article.ContentFilepath()
	if err = gh.CreateFile(ctx, msg, path, []byte(data)); err != nil {
		return err
	}
	return nil
}
