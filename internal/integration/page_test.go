package integration_test

import (
	"errors"
	"github.com/earthboundkid/assert"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/github"
	"github.com/spotlightpa/almanack/internal/services/index"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServicePublish(t *testing.T) {
	be := assert.FailNow(t)
	ctx := t.Context()
	almlog.UseTestLogger(t)

	dbhandle := createTestDB(t)

	tmp := t.ArtifactDir()
	svc := almsvc.Services{
		DB:           dbhandle,
		Queries:      dbhandle.Queries(),
		ContentStore: github.NewMockClient(tmp),
		Indexer:      index.MockIndexer{},
	}

	// Success case
	{
		const path1 = "content/news/1.md"
		p0, err := svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path1,
			SourceType: "manual",
			SourceID:   "n/a",
		})
		be.Zero(err)

		p, err := svc.Queries.GetPageByFilePath(ctx, path1)
		be.Zero(err)
		be.False(p.LastPublished.Valid)
		be.Equal(p.ID, p0.ID)

		_, err = os.Stat(filepath.Join(tmp, path1))
		be.ErrorIs(err, os.ErrNotExist)

		p1 := &db.Page{
			ID:            p0.ID,
			FilePath:      path1,
			Frontmatter:   map[string]any{},
			Body:          "hello",
			ScheduleFor:   pgtype.Timestamptz{},
			LastPublished: pgtype.Timestamptz{},
			CreatedAt:     time.Time{},
			UpdatedAt:     time.Time{},
			URLPath: pgtype.Text{
				String: "/hello", Valid: true,
			},
			SourceType:      "",
			SourceID:        "",
			PublicationDate: pgtype.Timestamptz{},
		}
		err = svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := svc.PublishPage(ctx, txq, p1)
			be.Zero(warning)
			return err
		})
		be.Zero(err)

		p, err = svc.Queries.GetPageByFilePath(ctx, path1)
		be.Zero(err)
		be.True(p.LastPublished.Valid)
	}
	{
		const path2 = "content/news/2.md"
		_, err := svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path2,
			SourceType: "manual",
			SourceID:   "n/a",
		})
		be.Zero(err)

		_, err = os.Stat(filepath.Join(tmp, path2))
		be.ErrorIs(err, os.ErrNotExist)

		// Can't create another page with the same URLPath
		p2 := &db.Page{
			ID:            1,
			FilePath:      path2,
			Frontmatter:   map[string]any{},
			Body:          "hello",
			ScheduleFor:   pgtype.Timestamptz{},
			LastPublished: pgtype.Timestamptz{},
			CreatedAt:     time.Time{},
			UpdatedAt:     time.Time{},
			URLPath: pgtype.Text{
				String: "/hello", Valid: true,
			},
			SourceType:      "",
			SourceID:        "",
			PublicationDate: pgtype.Timestamptz{},
		}
		err = svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := svc.PublishPage(ctx, txq, p2)
			be.Zero(warning)
			return err
		})
		be.NotZero(err)
		_, err = os.Stat(filepath.Join(tmp, path2))
		be.ErrorIs(err, os.ErrNotExist)

		// Can create if the URL changes
		p3 := &db.Page{
			ID:            1,
			FilePath:      path2,
			Frontmatter:   map[string]any{},
			Body:          "hello",
			ScheduleFor:   pgtype.Timestamptz{},
			LastPublished: pgtype.Timestamptz{},
			CreatedAt:     time.Time{},
			UpdatedAt:     time.Time{},
			URLPath: pgtype.Text{
				String: "/hello2", Valid: true,
			},
			SourceType:      "",
			SourceID:        "",
			PublicationDate: pgtype.Timestamptz{},
		}
		err = svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := svc.PublishPage(ctx, txq, p3)
			be.Zero(warning)
			return err
		})
		be.Zero(err)
		_, err = os.Stat(filepath.Join(tmp, path2))
		be.Zero(err)
	}
	// Test Github failure
	{
		const path3 = "content/news/3.md"
		_, err := svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path3,
			SourceType: "manual",
			SourceID:   "n/a",
		})
		be.Zero(err)

		_, err = os.Stat(filepath.Join(tmp, path3))
		be.ErrorIs(err, os.ErrNotExist)

		p4 := &db.Page{
			ID:            1,
			FilePath:      path3,
			Frontmatter:   map[string]any{},
			Body:          "hello",
			ScheduleFor:   pgtype.Timestamptz{},
			LastPublished: pgtype.Timestamptz{},
			CreatedAt:     time.Time{},
			UpdatedAt:     time.Time{},
			URLPath: pgtype.Text{
				String: "/hello3", Valid: true,
			},
			SourceType:      "",
			SourceID:        "",
			PublicationDate: pgtype.Timestamptz{},
		}
		// Github returns an error
		svc.ContentStore = github.ErrorClient{
			Error: errors.New("bad client"),
		}
		err = svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := svc.PublishPage(ctx, txq, p4)
			be.Zero(warning)
			return err
		})
		be.NotZero(err)

		p, err := svc.Queries.GetPageByFilePath(ctx, path3)
		be.Zero(err)
		be.False(p.LastPublished.Valid)
	}
}

func TestServicePublishTaxonomyPages(t *testing.T) {
	be := assert.FailNow(t)
	ctx := t.Context()
	almlog.UseTestLogger(t)

	dbhandle := createTestDB(t)

	tmp := t.ArtifactDir()
	svc := almsvc.Services{
		DB:           dbhandle,
		Queries:      dbhandle.Queries(),
		ContentStore: github.NewMockClient(tmp),
		Indexer:      index.MockIndexer{},
	}

	const storyPath = "content/news/taxo.md"

	p := &db.Page{
		FilePath:   storyPath,
		SourceType: "manual",
		SourceID:   "n/a",
		Frontmatter: db.Map{
			"topics":      []any{"Health", "Education"},
			"series":      []any{"Capitol Notebook"},
			"description": "a desc",
			"image":       "img.jpg",
		},
		Body: "hello",
	}

	err := svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
		txerr = p.Save(ctx, txq, false)
		be.Zero(txerr)

		err, warning := svc.PublishPage(ctx, txq, p)
		be.Zero(warning)
		return err
	})
	be.Zero(err)

	// Source page was published.
	_, err = os.Stat(filepath.Join(tmp, storyPath))
	be.Zero(err)

	// Taxonomy pages were created in the DB and in the content store.
	wantPaths := []string{
		"content/topics/Health/_index.md",
		"content/topics/Education/_index.md",
		"content/series/Capitol Notebook/_index.md",
	}
	for _, path := range wantPaths {
		tp, err := svc.Queries.GetPageByFilePath(ctx, path)
		be.Zero(err)
		be.Equal(tp.SourceType, "taxonomy")
		be.Equal(tp.SourceID, storyPath)
		be.True(tp.LastPublished.Valid)
		be.NotZero(tp.URLPath)
		_, err = os.Stat(filepath.Join(tmp, path))
		be.Zero(err)
	}

	// Republishing the page should not produce duplicate taxonomy pages
	// or error out.
	err = svc.DB.Tx(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
		err, warning := svc.PublishPage(ctx, txq, p)
		be.Zero(warning)
		return err
	})
	be.Zero(err)
}

func TestServicePopScheduledPages(t *testing.T) {
	be := assert.FailNow(t)
	ctx := t.Context()
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)

	tmp := t.ArtifactDir()
	svc := almsvc.Services{
		DB:           dbhandle,
		Queries:      dbhandle.Queries(),
		ContentStore: github.NewMockClient(tmp),
		Indexer:      index.MockIndexer{},
	}

	{
		const path = "content/news/test-pop.md"
		p, err := svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path,
			SourceType: "manual",
			SourceID:   "n/a",
		})
		be.Zero(err)

		p, err = svc.Queries.GetPageByFilePath(ctx, path)
		be.Zero(err)
		be.False(p.LastPublished.Valid)

		_, err = os.Stat(filepath.Join(tmp, path))
		be.ErrorIs(err, os.ErrNotExist)

		p, err = svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
			ID:             p.ID,
			SetFrontmatter: false,
			Frontmatter:    map[string]any{},
			SetBody:        false,
			Body:           "",
			SetScheduleFor: true,
			ScheduleFor: pgtype.Timestamptz{
				Time:  time.Now().AddDate(0, 0, -1),
				Valid: true,
			},
			URLPath:          "",
			SetLastPublished: false,
		})
		be.Zero(err)
		be.False(p.LastPublished.Valid)

		err, warning := svc.PopScheduledPages(ctx)
		be.Zero(warning)
		be.Zero(err)

		p, err = svc.Queries.GetPageByFilePath(ctx, path)
		be.Zero(err)
		be.True(p.LastPublished.Valid)

		_, err = os.Stat(filepath.Join(tmp, path))
		be.Zero(err)
	}
}
