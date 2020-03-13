package almanack

import (
	"encoding/json"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
)

type ScheduledArticle struct {
	Article
	LastArcSync   time.Time
	ScheduleFor   *time.Time
	LastSaved     *time.Time
	LastPublished *time.Time
}

func (schArticle *ScheduledArticle) toSPLData() interface{} {
	type splDataType struct {
		ID               *string     `json:"internal-id"`
		Budget           *string     `json:"budget"`
		ImageURL         *string     `json:"image-url"`
		ImageCaption     *string     `json:"image-caption"`
		ImageCredit      *string     `json:"image-credit"`
		PubDate          *time.Time  `json:"pub-date"`
		Slug             *string     `json:"slug"`
		Authors          *[]string   `json:"authors"`
		Byline           *string     `json:"byline"`
		Hed              *string     `json:"hed"`
		Subhead          *string     `json:"subhead"`
		Summary          *string     `json:"summary"`
		Blurb            *string     `json:"blurb"`
		Kicker           *string     `json:"kicker"`
		Body             *string     `json:"body"`
		LinkTitle        *string     `json:"link-title"`
		SuppressFeatured *bool       `json:"suppress-featured"`
		LastArcSync      *time.Time  `json:"last-arc-sync"`
		LastSaved        **time.Time `json:"last-saved"`
	}
	return &splDataType{
		ID:               &schArticle.InternalID,
		Budget:           &schArticle.Budget,
		ImageURL:         &schArticle.ImageURL,
		ImageCaption:     &schArticle.ImageCaption,
		ImageCredit:      &schArticle.ImageCredit,
		PubDate:          &schArticle.PubDate,
		Slug:             &schArticle.Slug,
		Authors:          &schArticle.Authors,
		Byline:           &schArticle.Byline,
		Hed:              &schArticle.Hed,
		Subhead:          &schArticle.Subhead,
		Summary:          &schArticle.Summary,
		Blurb:            &schArticle.Blurb,
		Kicker:           &schArticle.Kicker,
		Body:             &schArticle.Body,
		LinkTitle:        &schArticle.LinkTitle,
		SuppressFeatured: &schArticle.SuppressFeatured,
		LastArcSync:      &schArticle.LastArcSync,
		LastSaved:        &schArticle.LastSaved,
	}
}

func (schArticle *ScheduledArticle) fromDB(dbArticle db.Article) error {
	schArticle.ArcID = dbArticle.ArcID.String
	schArticle.ScheduleFor = timeNull(dbArticle.ScheduleFor)
	schArticle.LastPublished = timeNull(dbArticle.LastPublished)
	schArticle.filepath = dbArticle.SpotlightPAPath.String

	if err := json.Unmarshal(dbArticle.SpotlightPAData, schArticle.toSPLData()); err != nil {
		return err
	}

	if schArticle.LastArcSync.IsZero() {
		if err := schArticle.ResetArcData(dbArticle); err != nil {
			return err
		}
	}
	return nil
}

func (schArticle *ScheduledArticle) ResetArcData(dbArticle db.Article) error {
	schArticle.LastArcSync = time.Now()
	var arcStory ArcStory
	if err := json.Unmarshal(dbArticle.ArcData, &arcStory); err != nil {
		return err
	}
	art, err := arcStory.ToArticle()
	if err != nil {
		return err
	}
	schArticle.Article = *art
	return nil
}

func (schArticle *ScheduledArticle) toDB() (*db.Article, error) {
	var dart db.Article

	dart.ArcID = nullString(schArticle.ArcID)
	dart.ScheduleFor = nullTime(schArticle.ScheduleFor)
	dart.LastPublished = nullTime(schArticle.LastPublished)
	dart.SpotlightPAPath = nullString(schArticle.ContentFilepath())

	var err error
	if dart.SpotlightPAData, err = json.Marshal(schArticle.toSPLData()); err != nil {
		return nil, err
	}
	return &dart, nil
}
