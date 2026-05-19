package db_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
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
			err = p2.FromMD(toml)
			be.NilErr(t, err)
			be.Equal(t, fmt.Sprint(p1), fmt.Sprint(p2))
		})
	}
}

func TestFromToTOML(t *testing.T) {
	testfile.Run(t, "testdata/*.md", func(t *testing.T, path string) {
		s := testfile.Read(t, path)

		var page db.Page
		err := page.FromMD(s)
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
		"statecollege-date": {
			db.Page{
				FilePath: "content/statecollege/foo.md",
				Frontmatter: db.Map{
					// Mid-year UTC instant still resolves to the same EST month.
					"published": time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC),
				},
			},
			"/statecollege/2023/06/foo/",
		},
		"berks-date": {
			db.Page{
				FilePath: "content/berks/foo.md",
				Frontmatter: db.Map{
					"published": time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC),
				},
			},
			"/berks/2023/06/foo/",
		},
		"sponsored-date": {
			db.Page{
				FilePath: "content/sponsored/foo.md",
				Frontmatter: db.Map{
					"published": time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC),
				},
			},
			"/sponsored/2023/06/foo/",
		},
		"topics-index": {
			// Section landing pages (_index.md) keep the bare directory.
			db.Page{FilePath: "content/topics/elections/_index.md"},
			"/topics/elections/",
		},
		"series-index": {
			// Names get slugified
			db.Page{FilePath: "content/series/An Investigation/_index.md"},
			"/series/an-investigation/",
		},
		"index-slug": {
			// Slug+_index.md
			db.Page{
				FilePath: "content/series/An Investigation/_index.md",
				Frontmatter: db.Map{
					"slug": "investigation",
				},
			},
			"/series/investigation/",
		},
		"url-uppercased": {
			db.Page{
				FilePath: "content/abc/123.md",
				Frontmatter: db.Map{
					"url": "/Hello-World/",
				},
			},
			"/hello-world/",
		},
		"url-blank-fallback": {
			// An empty url falls through to the slug/path logic.
			db.Page{
				FilePath: "content/abc/123.md",
				Frontmatter: db.Map{
					"url":  "",
					"slug": "xyz",
				},
			},
			"/abc/xyz/",
		},
		"already-set-blank": {
			// A URLPath flagged Valid but empty is treated as unset.
			db.Page{
				FilePath: "content/abc/123.md",
				URLPath:  pgtype.Text{String: "", Valid: true},
			},
			"/abc/123/",
		},
		"nested-path": {
			db.Page{FilePath: "content/projects/series-a/part-1.md"},
			"/projects/series-a/part-1/",
		},
		"news-pubdate-jan-est-rolls-back": {
			// Jan 1 UTC midnight is Dec 31 EST, so the month should roll back.
			db.Page{
				FilePath: "content/news/abc.md",
				Frontmatter: db.Map{
					"published": time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			"/news/2019/12/abc/",
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
