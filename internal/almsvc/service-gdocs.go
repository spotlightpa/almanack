package almsvc

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/gdocs"
	"github.com/spotlightpa/almanack/internal/utils/must"
)

func (svc Services) CreateGDocsDoc(ctx context.Context, externalID string) (dbDoc *db.GDocsDoc, err error) {
	defer errorx.Trace(&err)

	if err := svc.ConfigureGoogleCert(ctx); err != nil {
		return nil, err
	}
	cl, err := svc.Gsvc.GDocsClient(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := gdocs.Request(ctx, cl, externalID)
	if err != nil {
		// TODO: figure out common errors, like no-permissions
		return nil, err
	}

	newDoc, err := svc.Queries.CreateGDocsDoc(ctx, db.CreateGDocsDocParams{
		ExternalID: externalID,
		Document:   *doc,
	})
	if err != nil {
		return nil, err
	}
	return &newDoc, nil
}

func (svc Services) ProcessGDocs(ctx context.Context) error {
	docs, err := svc.Queries.ListGDocsWhereUnprocessed(ctx)
	if err != nil {
		return err
	}

	var errs []error
	for _, doc := range docs {
		docErr := svc.ProcessGDocsDoc(ctx, doc)
		errs = append(errs, docErr)
	}
	return errors.Join(errs...)
}

func (svc Services) InflateSharedArticle(ctx context.Context, a *db.SharedArticle) (v any, err error) {
	type GDocsWithoutDocument struct {
		*db.GDocsDoc
		Document string `json:"document,omitempty"`
	}

	type SharedArticleWithGDoc struct {
		*db.SharedArticle
		GDocs GDocsWithoutDocument `json:"gdocs"`
	}
	defer errorx.Trace(&err)

	if a.SourceType != "gdocs" {
		return a, nil
	}
	var id int64
	if err = json.Unmarshal(a.RawData, &id); err != nil {
		return nil, err
	}

	doc, err := svc.Queries.GetGDocsByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return SharedArticleWithGDoc{a, GDocsWithoutDocument{&doc, ""}}, err
}

func (svc Services) UpsertSharedArticleForGDoc(ctx context.Context, dbDoc *db.GDocsDoc, refreshMetadata bool) (*db.SharedArticle, error) {
	idJSON := must.Get(json.Marshal(dbDoc.ID))
	if refreshMetadata {
		art, err := svc.Queries.UpdateSharedArticleFromGDocs(ctx, db.UpdateSharedArticleFromGDocsParams{
			ExternalID:           dbDoc.ExternalID,
			RawData:              idJSON,
			InternalID:           dbDoc.Metadata.InternalID,
			Byline:               dbDoc.Metadata.Byline,
			Budget:               dbDoc.Metadata.Budget,
			Hed:                  dbDoc.Metadata.Hed,
			Description:          dbDoc.Metadata.Description,
			Blurb:                dbDoc.Metadata.Blurb,
			LedeImage:            dbDoc.Metadata.LedeImage,
			LedeImageCredit:      dbDoc.Metadata.LedeImageCredit,
			LedeImageDescription: dbDoc.Metadata.LedeImageDescription,
			LedeImageCaption:     dbDoc.Metadata.LedeImageCaption,
		})
		return &art, err
	}
	art, err := svc.Queries.UpsertSharedArticleFromGDocs(ctx, db.UpsertSharedArticleFromGDocsParams{
		ExternalID:           dbDoc.ExternalID,
		RawData:              idJSON,
		InternalID:           dbDoc.Metadata.InternalID,
		Byline:               dbDoc.Metadata.Byline,
		Budget:               dbDoc.Metadata.Budget,
		Hed:                  dbDoc.Metadata.Hed,
		Description:          dbDoc.Metadata.Description,
		Blurb:                dbDoc.Metadata.Blurb,
		LedeImage:            dbDoc.Metadata.LedeImage,
		LedeImageCredit:      dbDoc.Metadata.LedeImageCredit,
		LedeImageDescription: dbDoc.Metadata.LedeImageDescription,
		LedeImageCaption:     dbDoc.Metadata.LedeImageCaption,
	})
	return &art, err
}
