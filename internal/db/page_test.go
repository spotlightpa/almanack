package db_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func TestToFromTOML(t *testing.T) {
	cases := map[string]db.Page{
		"empty": {Frontmatter: db.Map{}},
		"body":  {Frontmatter: db.Map{}, Body: "\n ## subhead ! \n"},
		"fm": {Frontmatter: db.Map{
			"Hed":     "Hello",
			"Authors": []string{"john", "smith"}}},
		"body+fm": {
			Frontmatter: db.Map{
				"Hed":     "Hello",
				"N":       1,
				"Authors": []string{"john", "smith"}},
			Body: "## subhead !"},
		"extra-delimiters": {
			Frontmatter: db.Map{
				"Hed":     "Hello",
				"Authors": []string{"john", "smith"}},
			Body: "## subhead !\n+++\n\nmore\n+++\nstuff",
		},
	}
	for name, p1 := range cases {
		t.Run(name, func(t *testing.T) {
			toml, err := p1.ToTOML()
			be.NilErr(t, err)

			var p2 db.Page
			err = p2.FromTOML(toml)
			be.NilErr(t, err)
			be.Equal(t, fmt.Sprint(p1), fmt.Sprint(p2))
		})
	}
}

func TestFromToTOML(t *testing.T) {
	testfile.Run(t, "testdata/*.md", func(t *testing.T, path string) {
		b := testfile.Read(t, path)

		var page db.Page
		err := page.FromTOML(string(b))
		be.NilErr(t, err)

		toml, err := page.ToTOML()
		be.NilErr(t, err)

		testfile.Equal(t, path, toml)
	})
}

func TestSetURLPath(t *testing.T) {
	cases := map[string]struct {
		db.Page
		string
	}{
		"blank": {
			db.Page{}, "",
		},
		"no-slug": {
			db.Page{FilePath: "content/abc/123.md"}, "/abc/123/",
		},
		"already-set": {
			db.Page{
				FilePath: "content/abc/123.md",
				URLPath: pgtype.Text{
					String: "/xyz/345/",
					Valid:  true,
				},
			},
			"/xyz/345/",
		},
		"from-slug": {
			db.Page{
				FilePath: "content/abc/123.md",
				Frontmatter: db.Map{
					"slug": "567",
				},
			},
			"/abc/567/",
		},
		"from-url": {
			db.Page{
				FilePath: "content/abc/123.md",
				Frontmatter: db.Map{
					"slug": "567",
					"url":  "/hello-world",
				},
			},
			"/hello-world",
		},
		"news-date": {
			db.Page{
				FilePath: "content/news/123.md",
				Frontmatter: db.Map{
					"published": time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC),
				},
			},
			"/news/2019/12/123/",
		},
		"news-date-string": {
			db.Page{
				FilePath: "content/news/123.md",
				Frontmatter: db.Map{
					"published": "2020-01-01T00:00:00.000Z",
				},
			},
			"/news/2019/12/123/",
		},
		"news-date-slug": {
			db.Page{
				FilePath: "content/news/123.md",
				Frontmatter: db.Map{
					"published": time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC),
					"slug":      "abc",
				},
			},
			"/news/2019/12/abc/",
		},
		"news-url": {
			db.Page{
				FilePath: "content/news/123.md",
				Frontmatter: db.Map{
					"published": time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC),
					"slug":      "abc",
					"url":       "/hello-world",
				},
			},
			"/hello-world",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.Page.SetURLPath()
			be.Equal(t,
				tc.Page.URLPath.Valid,
				tc.Page.URLPath.String != "")
			be.Equal(t, tc.string, tc.Page.URLPath.String)
		})
	}
}

func TestShouldPublishShouldNotify(t *testing.T) {
	past := pgtype.Timestamptz{
		Valid: true}
	future := pgtype.Timestamptz{
		Valid: true,
		Time:  time.Now().Add(24 * time.Hour)}
	cases := map[string]struct {
		old, new    db.Page
		pub, notify bool
	}{
		"blank": {
			old:    db.Page{},
			new:    db.Page{},
			pub:    false,
			notify: false,
		},
		"scheduled-news": {
			old: db.Page{
				FilePath: "content/news/whatever.md"},
			new: db.Page{
				FilePath:    "content/news/whatever.md",
				ScheduleFor: future},
			pub:    false,
			notify: true,
		},
		"scheduled-statecollege": {
			old: db.Page{
				FilePath: "content/statecollege/whatever.md"},
			new: db.Page{
				FilePath:    "content/statecollege/whatever.md",
				ScheduleFor: future},
			pub:    false,
			notify: true,
		},
		"rescheduled-news": {
			old: db.Page{
				FilePath:    "content/news/whatever.md",
				ScheduleFor: past},
			new: db.Page{
				FilePath:    "content/news/whatever.md",
				ScheduleFor: future},
			pub:    false,
			notify: true,
		},
		"pub-news": {
			old: db.Page{
				FilePath: "content/news/whatever.md"},
			new: db.Page{
				FilePath:    "content/news/whatever.md",
				ScheduleFor: past},
			pub:    true,
			notify: true,
		},
		"pub-statecollege": {
			old: db.Page{
				FilePath: "content/statecollege/whatever.md"},
			new: db.Page{
				FilePath:    "content/statecollege/whatever.md",
				ScheduleFor: past},
			pub:    true,
			notify: true,
		},
		"repub-news": {
			old: db.Page{
				FilePath:      "content/news/whatever.md",
				LastPublished: past},
			new: db.Page{
				FilePath:    "content/news/whatever.md",
				ScheduleFor: past},
			pub:    true,
			notify: false,
		},
		"scheduled-non-news": {
			old: db.Page{},
			new: db.Page{
				ScheduleFor: future},
			pub:    false,
			notify: false,
		},
		"rescheduled-non-news": {
			old: db.Page{
				ScheduleFor: past},
			new: db.Page{
				ScheduleFor: future},
			pub:    false,
			notify: false,
		},
		"pub-non-news": {
			old: db.Page{},
			new: db.Page{
				ScheduleFor: past},
			pub:    true,
			notify: false,
		},
		"repub-non-news": {
			old: db.Page{
				LastPublished: past},
			new: db.Page{
				ScheduleFor: past},
			pub:    true,
			notify: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			pub := tc.new.ShouldPublish()
			notify := tc.new.ShouldNotify(&tc.old)
			be.Equal(t, tc.pub, pub)
			be.Equal(t, tc.notify, notify)
		})
	}
}

func TestServicePublish(t *testing.T) {
	ctx := context.Background()
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

		be.NilErr(t, svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path1,
			SourceType: "manual",
			SourceID:   "n/a",
		}))

		p, err := svc.Queries.GetPageByFilePath(ctx, path1)
		be.NilErr(t, err)
		be.False(t, p.LastPublished.Valid)

		_, err = os.Stat(filepath.Join(tmp, path1))
		be.Nonzero(t, err)

		p1 := &db.Page{
			ID:            1,
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

		be.NilErr(t, svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path2,
			SourceType: "manual",
			SourceID:   "n/a",
		}))

		_, err := os.Stat(filepath.Join(tmp, path2))
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
		be.NilErr(t, svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path3,
			SourceType: "manual",
			SourceID:   "n/a",
		}))

		_, err := os.Stat(filepath.Join(tmp, path3))
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
	ctx := context.Background()
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

		be.NilErr(t, svc.Queries.CreatePage(ctx, db.CreatePageParams{
			FilePath:   path,
			SourceType: "manual",
			SourceID:   "n/a",
		}))

		p, err := svc.Queries.GetPageByFilePath(ctx, path)
		be.NilErr(t, err)
		be.False(t, p.LastPublished.Valid)

		_, err = os.Stat(filepath.Join(tmp, path))
		be.Nonzero(t, err)

		p, err = svc.Queries.UpdatePage(ctx, db.UpdatePageParams{
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
			FilePath:         path,
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
