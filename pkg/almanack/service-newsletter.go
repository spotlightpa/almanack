package almanack

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/carlmjohnson/errorx"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/internal/timex"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (svc Services) UpdateNewsletterArchives(ctx context.Context, types []db.NewsletterType) (err error) {
	campaigns, err := svc.NewletterService.ListCampaigns(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, nltype := range types {
		errs = append(errs, svc.UpdateNewsletterArchive(ctx, campaigns, nltype.Name, nltype.Shortname))
	}
	return errors.Join(errs...)
}

func (svc Services) UpdateNewsletterArchive(ctx context.Context, campaigns *mailchimp.ListCampaignsResp, mcType, dbType string) (err error) {
	defer errorx.Trace(&err)

	newItems := campaigns.ToNewsletters(mcType)
	// update DB
	data, err := json.Marshal(newItems)
	if err != nil {
		return err
	}
	n, err := svc.Queries.UpdateNewsletterArchives(ctx, db.UpdateNewsletterArchivesParams{
		Type: dbType,
		Data: data,
	})
	if err != nil {
		return err
	}
	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "Services.UpdateNewsletterArchive",
		"mcType", mcType, "new-items", n)

	return nil
}

func (svc Services) ImportNewsletterPages(ctx context.Context, types []db.NewsletterType) (err error) {
	defer errorx.Trace(&err)

	nls, err := svc.Queries.ListNewslettersWithoutPage(ctx, db.ListNewslettersWithoutPageParams{
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		return err
	}
	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "Services.ImportNewsletterPages",
		"new-items", len(nls))

	for _, nl := range nls {
		body, err := mailchimp.ImportPage(ctx, svc.Client, nl.ArchiveURL)
		if err != nil {
			return err
		}
		if err = svc.SaveNewsletterPage(ctx, &nl, body, types); err != nil {
			return err
		}
	}
	return nil
}

func (svc Services) SaveNewsletterPage(ctx context.Context, nl *db.Newsletter, body string, types []db.NewsletterType) (err error) {
	defer errorx.Trace(&err)

	needsUpdate := false
	if nl.SpotlightPAPath.String == "" {
		nl.PublishedAt = timex.ToEST(nl.PublishedAt)
		nl.SpotlightPAPath.Valid = true
		nl.SpotlightPAPath.String = fmt.Sprintf("content/newsletters/%s/%s.md",
			nl.Type, nl.PublishedAt.Format("2006-01-02-1504"),
		)
		needsUpdate = true
	}

	// create or update the page
	if !needsUpdate {
		return nil
	}
	path := nl.SpotlightPAPath.String
	if err := svc.Queries.CreatePage(ctx, db.CreatePageParams{
		FilePath:   path,
		SourceType: "mailchimp",
		SourceID:   strconv.FormatInt(nl.ID, 10),
	}); err != nil {
		return err
	}
	slug := stringx.SlugifyURL(
		timex.ToEST(nl.PublishedAt).Format("Jan 2 ") + nl.Subject,
	)
	kicker := "Newsletter"
	for _, nltype := range types {
		if nltype.Shortname == nl.Type {
			kicker = nltype.Name
		}
	}
	if _, err := svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
		SetFrontmatter: true,
		Frontmatter: map[string]any{
			"aliases":           []string{},
			"authors":           []string{},
			"blurb":             nl.Blurb,
			"byline":            "",
			"description":       nl.Description,
			"draft":             false,
			"extended-kicker":   "",
			"image":             "",
			"image-caption":     "",
			"image-credit":      "",
			"image-description": "",
			"image-size":        "",
			"internal-id": fmt.Sprintf("%s-%s",
				strings.ToUpper(nl.Type),
				nl.PublishedAt.Format("01-02-06")),
			"kicker":      kicker,
			"layout":      "mailchimp-page",
			"linktitle":   "",
			"no-index":    false,
			"published":   nl.PublishedAt,
			"raw-content": body,
			"series":      []string{},
			"slug":        slug,
			"title":       nl.Subject,
			"title-tag":   "",
			"topics":      []string{},
			"url":         "",
		},
		SetBody:     true,
		Body:        "",
		FilePath:    path,
		ScheduleFor: db.NullTime,
	}); err != nil {
		return err
	}

	if nl2, err := svc.Queries.SetNewsletterPage(ctx, db.SetNewsletterPageParams{
		ID:              nl.ID,
		SpotlightPAPath: nl.SpotlightPAPath,
	}); err != nil {
		return err
	} else {
		*nl = nl2
	}

	return nil
}
