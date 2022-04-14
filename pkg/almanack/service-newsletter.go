package almanack

import (
	"context"
	"fmt"
	"strings"

	"github.com/carlmjohnson/errutil"
	"github.com/jackc/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/stringutils"
	"github.com/spotlightpa/almanack/internal/timeutil"
)

func (svc Service) UpdateNewsletterArchives(ctx context.Context) error {
	return errutil.ExecParallel(
		func() error {
			return svc.UpdateNewsletterArchive(
				ctx,
				"The Investigator",
				"investigator",
			)
		},
		func() error {
			return svc.UpdateNewsletterArchive(
				ctx,
				"PA Post",
				"papost",
			)
		},
		func() error {
			return svc.UpdateNewsletterArchive(
				ctx,
				"PA Local",
				"palocal",
			)
		},
	)
}

func (svc Service) UpdateNewsletterArchive(ctx context.Context, mcType, dbType string) (err error) {
	defer errutil.Trace(&err)

	// get the latest from MC
	newItems, err := svc.NewletterService.ListNewletters(ctx, mcType)
	if err != nil {
		return err
	}
	// update DB
	var data pgtype.JSONB
	if err = data.Set(newItems); err != nil {
		return err
	}
	if n, err := svc.Queries.UpdateNewsletterArchives(ctx, db.UpdateNewsletterArchivesParams{
		Type: dbType,
		Data: data,
	}); err != nil {
		return err
	} else if n == 0 {
		// abort if there's nothing new to update
		svc.Logger.Printf("%q got no new items", mcType)
		return nil
	} else {
		svc.Logger.Printf("%q got %d new items", mcType, n)
	}

	return nil
}

func (svc Service) ImportNewsletterPages(ctx context.Context) (err error) {
	defer errutil.Trace(&err)

	nls, err := svc.Queries.ListNewslettersWithoutPage(ctx, db.ListNewslettersWithoutPageParams{
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		return err
	}
	svc.Logger.Printf("importing %d newsletter pages", len(nls))
	for _, nl := range nls {
		body, err := mailchimp.ImportPage(ctx, svc.Client, nl.ArchiveURL)
		if err != nil {
			return err
		}
		if err = svc.SaveNewsletterPage(ctx, &nl, body); err != nil {
			return err
		}
	}
	return nil
}

var kickerFor = map[string]string{
	"investigator": "The Investigator",
	"papost":       "PA Post",
	"palocal":      "PA Local",
}

func (svc Service) SaveNewsletterPage(ctx context.Context, nl *db.Newsletter, body string) (err error) {
	defer errutil.Trace(&err)

	needsUpdate := false
	if nl.SpotlightPAPath.String == "" {
		nl.PublishedAt = timeutil.ToEST(nl.PublishedAt)
		nl.SpotlightPAPath.Status = pgtype.Present
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
	if err := svc.Queries.EnsurePage(ctx, path); err != nil {
		return err
	}
	slug := stringutils.Slugify(
		timeutil.ToEST(nl.PublishedAt).Format("Jan 2 ") + nl.Subject,
	)
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
			//TODO: proper kicker lookup
			"kicker":      kickerFor[nl.Type],
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
