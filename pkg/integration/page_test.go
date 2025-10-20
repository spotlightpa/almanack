package db_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestServicePublish(t *testing.T) {
	ctx := t.Context()
	almlog.UseTestLogger(t)

	p := createTestDB(t)
	tmp := t.TempDir()
	svc := almanack.Services{
		Queries:      db.New(p),
		Tx:           db.NewTxable(p),
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
		be.NilErr(t, err)

		p, err := svc.Queries.GetPageByFilePath(ctx, path1)
		be.NilErr(t, err)
		be.False(t, p.LastPublished.Valid)
		be.Equal(t, p0.ID, p.ID)

		_, err = os.Stat(filepath.Join(tmp, path1))
		be.Nonzero(t, err)

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
		err = svc.Tx.Begin(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := svc.PublishPage(ctx, txq, p1)
			be.NilErr(t, warning)
			return err
		})
		be.NilErr(t, err)

		p, err = svc.Queries.GetPageByFilePath(ctx, path1)
		be.NilErr(t, err)
		be.True(t, p.LastPublished.Valid)
	}
	{
		const path2 = "content/news/2.md"
		_, err := svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path2,
			SourceType: "manual",
			SourceID:   "n/a",
		})
		be.NilErr(t, err)

		_, err = os.Stat(filepath.Join(tmp, path2))
		be.Nonzero(t, err)

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
		err = svc.Tx.Begin(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := svc.PublishPage(ctx, txq, p2)
			be.NilErr(t, warning)
			return err
		})
		be.Nonzero(t, err)
		_, err = os.Stat(filepath.Join(tmp, path2))
		be.Nonzero(t, err)

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
		err = svc.Tx.Begin(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := svc.PublishPage(ctx, txq, p3)
			be.NilErr(t, warning)
			return err
		})
		be.NilErr(t, err)
		_, err = os.Stat(filepath.Join(tmp, path2))
		be.NilErr(t, err)
	}
	// Test Github failure
	{
		const path3 = "content/news/3.md"
		_, err := svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path3,
			SourceType: "manual",
			SourceID:   "n/a",
		})
		be.NilErr(t, err)

		_, err = os.Stat(filepath.Join(tmp, path3))
		be.Nonzero(t, err)

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
		err = svc.Tx.Begin(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := svc.PublishPage(ctx, txq, p4)
			be.NilErr(t, warning)
			return err
		})
		be.Nonzero(t, err)

		p, err := svc.Queries.GetPageByFilePath(ctx, path3)
		be.NilErr(t, err)
		be.False(t, p.LastPublished.Valid)
	}
}

func TestServicePopScheduledPages(t *testing.T) {
	ctx := t.Context()
	almlog.UseTestLogger(t)

	p := createTestDB(t)
	tmp := t.TempDir()
	svc := almanack.Services{
		Queries:      db.New(p),
		Tx:           db.NewTxable(p),
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
		be.NilErr(t, err)

		p, err = svc.Queries.GetPageByFilePath(ctx, path)
		be.NilErr(t, err)
		be.False(t, p.LastPublished.Valid)

		_, err = os.Stat(filepath.Join(tmp, path))
		be.Nonzero(t, err)

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
		be.NilErr(t, err)
		be.False(t, p.LastPublished.Valid)

		err, warning := svc.PopScheduledPages(ctx)
		be.NilErr(t, warning)
		be.NilErr(t, err)

		p, err = svc.Queries.GetPageByFilePath(ctx, path)
		be.NilErr(t, err)
		be.True(t, p.LastPublished.Valid)

		_, err = os.Stat(filepath.Join(tmp, path))
		be.NilErr(t, err)
	}
}
