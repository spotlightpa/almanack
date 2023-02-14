package db_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/carlmjohnson/be"
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
			err = p2.FromTOML(toml)
			be.NilErr(t, err)
			be.Equal(t, fmt.Sprint(p1), fmt.Sprint(p2))
		})
	}
}

func TestFromToTOML(t *testing.T) {
	cases := []string{
		"blank.md",
		"fm.md",
		"fm+body.md",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile("testdata/" + name)
			be.NilErr(t, err)

			var page db.Page
			err = page.FromTOML(string(b))
			be.NilErr(t, err)

			toml, err := page.ToTOML()
			be.NilErr(t, err)

			be.Debug(t, func() {
				t.Errorf("%q did not round trip", name)
				os.WriteFile("testdata/bad-"+name, []byte(toml), 0644)
			})
			be.Equal(t, string(b), toml)
		})
	}
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
