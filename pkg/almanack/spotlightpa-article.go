package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/carlmjohnson/errutil"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/slack"
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
	SuppressFeatured bool       `toml:"suppress-featured"`
	Weight           int        `toml:"weight"`
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
}

func (splArt *SpotlightPAArticle) toSPLData() interface{} {
	type splDataType struct {
		ArcID            string     `json:"-"`
		InternalID       string     `json:"internal-id"`
		Budget           string     `json:"budget"`
		ImageURL         string     `json:"image-url"`
		ImageDescription string     `json:"image-description"`
		ImageCaption     string     `json:"image-caption"`
		ImageCredit      string     `json:"image-credit"`
		ImageSize        string     `json:"image-size"`
		PubDate          time.Time  `json:"pub-date"`
		Slug             string     `json:"slug"`
		Authors          []string   `json:"authors"`
		Byline           string     `json:"byline"`
		Hed              string     `json:"hed"`
		Subhead          string     `json:"subhead"`
		Summary          string     `json:"summary"`
		Blurb            string     `json:"blurb"`
		Kicker           string     `json:"kicker"`
		Topics           []string   `json:"topics"`
		Series           []string   `json:"series"`
		LinkTitle        string     `json:"link-title"`
		SuppressFeatured bool       `json:"suppress-featured"`
		Weight           int        `json:"weight"`
		OverrideURL      string     `json:"override-url"`
		Aliases          []string   `json:"aliases"`
		ModalExclude     bool       `json:"modal-exclude"`
		NoIndex          bool       `json:"no-index"`
		LanguageCode     string     `json:"language-code"`
		LayoutType       string     `json:"layout"`
		ExtendedKicker   string     `json:"extended-kicker"`
		Body             string     `json:"body"`
		Filepath         string     `json:"-"`
		LastArcSync      time.Time  `json:"last-arc-sync"`
		ScheduleFor      *time.Time `json:"-"`
		LastSaved        *time.Time `json:"last-saved"`
		LastPublished    *time.Time `json:"-"`
		Warnings         []string   `json:"-"`
	}
	return (*splDataType)(splArt)
}

func (splArt *SpotlightPAArticle) fromDB(dbArticle db.Article) error {
	splArt.ArcID = dbArticle.ArcID.String
	splArt.ScheduleFor = timeNull(dbArticle.ScheduleFor)
	splArt.LastPublished = timeNull(dbArticle.LastPublished)
	splArt.Filepath = dbArticle.SpotlightPAPath.String

	if err := json.Unmarshal(dbArticle.SpotlightPAData, splArt.toSPLData()); err != nil {
		return err
	}

	return nil
}

func (splArt *SpotlightPAArticle) Empty() bool {
	return splArt.LastArcSync.IsZero()
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

func (splArt *SpotlightPAArticle) toDB() (*db.Article, error) {
	var dart db.Article

	dart.ArcID = nullString(splArt.ArcID)
	dart.ScheduleFor = nullTime(splArt.ScheduleFor)
	dart.LastPublished = nullTime(splArt.LastPublished)
	dart.SpotlightPAPath = nullString(splArt.ContentFilepath())

	var err error
	if dart.SpotlightPAData, err = json.Marshal(splArt.toSPLData()); err != nil {
		return nil, err
	}
	return &dart, nil
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
	year := splArt.PubDate.Year()
	month := splArt.PubDate.Month()
	return fmt.Sprintf(
		"https://www.spotlightpa.org/news/%d/%02d/%s/",
		year, month, splArt.Slug,
	)
}

func (splArt *SpotlightPAArticle) ContentFilepath() string {
	if splArt.Filepath == "" {
		date := splArt.PubDate.Format("2006-01-02")
		splArt.Filepath = fmt.Sprintf("content/news/%s-%s.md", date, splArt.InternalID)
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
	var frontmatter, body string

	if !strings.HasPrefix(content, delimiter) {
		return fmt.Errorf("could not parse frontmatter: no prefix delimiter")
	}
	frontmatter = content[len(delimiter):]
	if end := strings.Index(frontmatter, delimiter); end != -1 {
		body = frontmatter[end+len(delimiter):]
		frontmatter = frontmatter[:end]
	} else {
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

func (splArt *SpotlightPAArticle) Publish(ctx context.Context, svc Service) error {
	return errutil.ExecParallel(
		func() error {
			return splArt.publishContent(ctx, svc)
		}, func() error {
			return splArt.indexContent(ctx, svc)
		})
}

func (splArt *SpotlightPAArticle) publishContent(ctx context.Context, svc Service) error {
	data, err := splArt.ToTOML()
	if err != nil {
		return err
	}
	path := splArt.ContentFilepath()
	msg := fmt.Sprintf("Content: publishing %q", splArt.InternalID)
	if err = svc.ContentStore.UpdateFile(ctx, msg, path, []byte(data)); err != nil {
		return err
	}
	return nil
}

func (splArt *SpotlightPAArticle) indexContent(ctx context.Context, svc Service) error {
	_, err := svc.Indexer.SaveObject(splArt.ToIndex(), ctx)
	return err
}

func (splArt *SpotlightPAArticle) Notify(ctx context.Context, svc Service) error {
	const (
		green  = "#78bc20"
		yellow = "#ffcb05"
	)
	text := "New article publishing now…"
	color := green

	if splArt.ScheduleFor != nil {
		t := splArt.ScheduleFor.Local()
		newYork, err := time.LoadLocation("America/New_York")
		if err == nil {
			t = splArt.ScheduleFor.In(newYork)
		}
		text = t.Format("New article scheduled for Mon, Jan 2 at 3:04pm MST…")
		color = yellow
	}

	return svc.SlackClient.PostCtx(ctx, slack.Message{
		Text: text,
		Attachments: []slack.Attachment{
			{
				Color: color,
				Fallback: fmt.Sprintf("%s\n%s\n%s",
					splArt.Hed, splArt.Summary, splArt.URL()),
				Title:     splArt.Hed,
				TitleLink: splArt.URL(),
				Text: fmt.Sprintf(
					"%s\n%s",
					splArt.Summary, splArt.URL()),
			},
		},
	})
}

func (splArt *SpotlightPAArticle) RefreshFromContentStore(ctx context.Context, svc Service) {
	if splArt.LastPublished == nil {
		return
	}
	content, err := svc.ContentStore.GetFile(ctx, splArt.ContentFilepath())
	if err != nil {
		splArt.Warnings = append(splArt.Warnings, err.Error())
		return
	}
	if err = splArt.FromTOML(content); err != nil {
		splArt.Warnings = append(splArt.Warnings, err.Error())
		return
	}
}

func (splArt *SpotlightPAArticle) ToIndex() interface{} {
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
		splArt.URL(),
		splArt.URL(),
		splArt.InternalID,
		splArt.ImageURL,
		splArt.ImageDescription,
		splArt.ImageCaption,
		splArt.ImageCredit,
		splArt.ImageSize,
		splArt.PubDate,
		splArt.Slug,
		splArt.Authors,
		splArt.Byline,
		splArt.Hed,
		splArt.Subhead,
		splArt.Summary,
		splArt.Blurb,
		splArt.Kicker,
		splArt.Topics,
		splArt.Series,
		splArt.LinkTitle,
		splArt.Aliases,
		splArt.Body,
	}
}
