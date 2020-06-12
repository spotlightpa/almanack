package almanack

import (
	"context"
	"encoding/json"
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
	SuppressFeatured bool       `toml:"suppress-featured"`
	Weight           int        `toml:"weight"`
	OverrideURL      string     `toml:"url"`
	Aliases          []string   `toml:"aliases"`
	ModalExclude     bool       `toml:"modal-exclude"`
	NoIndex          bool       `toml:"no-index"`
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

func (splArt *SpotlightPAArticle) Publish(ctx context.Context, svc Service) error {
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
