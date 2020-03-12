package almanack

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/spotlightpa/almanack/internal/db"
)

type ScheduledArticle struct {
	Article
	ScheduleFor *time.Time
	LastArcSync time.Time
	LastSaved   *time.Time
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

func ScheduledArticleFromDB(dbArticle db.Article) (*ScheduledArticle, error) {
	var schArticle ScheduledArticle
	schArticle.ArcID = dbArticle.ArcID.String
	if dbArticle.ScheduleFor.Valid {
		t := dbArticle.ScheduleFor.Time
		schArticle.ScheduleFor = &t
	}

	if err := json.Unmarshal(dbArticle.SpotlightPAData, schArticle.toSPLData()); err != nil {
		return nil, err
	}

	if schArticle.LastArcSync.IsZero() {
		if err := schArticle.ResetArcData(dbArticle); err != nil {
			return nil, err
		}
	}
	return &schArticle, nil
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

func (schArticle *ScheduledArticle) ToDB() (*db.Article, error) {
	var dart db.Article

	dart.ArcID = nullString(schArticle.ArcID)
	dart.SpotlightPAPath = nullString(schArticle.ContentFilepath())
	dart.ScheduleFor = nullTime(schArticle.ScheduleFor)
	var err error
	if dart.SpotlightPAData, err = json.Marshal(schArticle.toSPLData()); err != nil {
		return nil, err
	}
	return &dart, nil
}

type ScheduledArticleService struct {
	ContentStore
	Logger
	Querier db.Querier
}

func (sas ScheduledArticleService) Get(ctx context.Context, articleID string) (*ScheduledArticle, error) {
	start := time.Now()
	dart, err := sas.Querier.GetArticle(ctx, nullString(articleID))
	sas.Logger.Printf("queried GetArticle in %v", time.Since(start))
	if err != nil {
		return nil, db.StandardizeErr(err)
	}
	return ScheduledArticleFromDB(dart)
}

func (sas ScheduledArticleService) Save(ctx context.Context, article *ScheduledArticle) error {
	// TODO: Make less racey
	if article.ScheduleFor != nil &&
		article.ScheduleFor.Before(time.Now().Add(5*time.Minute)) {
		article.ScheduleFor = nil
		if err := article.Publish(ctx, sas.ContentStore); err != nil {
			return err
		}
	}

	// Save the article
	now := time.Now()
	article.LastSaved = &now

	dart, err := article.ToDB()
	if err != nil {
		return err
	}

	start := time.Now()
	*dart, err = sas.Querier.UpdateSpotlightPAArticle(ctx, db.UpdateSpotlightPAArticleParams{
		ArcID:           dart.ArcID,
		SpotlightPAData: dart.SpotlightPAData,
		ScheduleFor:     dart.ScheduleFor,
		SpotlightPAPath: dart.SpotlightPAPath,
	})
	sas.Logger.Printf("queried UpdateSpotlightPAArticle in %v", time.Since(start))
	if err != nil {
		return err
	}
	updatedArticle, err := ScheduledArticleFromDB(*dart)
	if err != nil {
		return err
	}
	*article = *updatedArticle
	return nil
}

func (sas ScheduledArticleService) PopScheduledArticles(ctx context.Context, callback func([]*ScheduledArticle) error) error {
	start := time.Now()
	poppedArts, err := sas.Querier.PopScheduled(ctx)
	sas.Logger.Printf("queried PopScheduled in %v", time.Since(start))
	if err != nil {
		return err
	}
	overdueArts := make([]*ScheduledArticle, len(poppedArts))
	for i := range overdueArts {
		overdueArts[i], err = ScheduledArticleFromDB(poppedArts[i])
		if err != nil {
			return err
		}
	}
	// If the status of the article changed, fire callback then update the list
	if len(overdueArts) > 0 {
		if err := callback(overdueArts); err != nil {
			// TODO rollback
			return err
		}
	}
	return nil
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullTime(t *time.Time) sql.NullTime {
	if t != nil {
		return sql.NullTime{Time: *t, Valid: true}
	}
	return sql.NullTime{}
}
