package db_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgtype"
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
			if err != nil {
				t.Fatalf("err ToTOML: %v", err)
			}
			var p2 db.Page
			if err = p2.FromTOML(toml); err != nil {
				t.Fatalf("err FromTOML: %v", err)
			}
			if fmt.Sprint(p1) != fmt.Sprint(p2) {
				t.Log(p1, p1)
				t.Errorf("article did not round trip")
			}
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
			if err != nil {
				t.Fatal(err)
			}
			var page db.Page
			err = page.FromTOML(string(b))
			if err != nil {
				t.Fatalf("err FromTOML: %v", err)
			}

			toml, err := page.ToTOML()
			if err != nil {
				t.Fatalf("err ToTOML: %v", err)
			}
			if toml != string(b) {
				os.WriteFile("testdata/bad-"+name, []byte(toml), 0644)
				t.Errorf("%q did not round trip", name)
			}
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
					Status: pgtype.Present,
					String: "/xyz/345/",
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
			if tc.Page.URLPath.Status == pgtype.Present != (tc.Page.URLPath.String != "") {
				t.Fatalf("bad validity")
			}

			if tc.Page.URLPath.String != tc.string {
				t.Fatalf("got %v", tc.Page.URLPath)
			}
		})
	}
}

func TestShouldPublishShouldNotify(t *testing.T) {
	past := pgtype.Timestamptz{
		Status: pgtype.Present}
	future := pgtype.Timestamptz{
		Status: pgtype.Present,
		Time:   time.Now().Add(24 * time.Hour)}
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
			if pub != tc.pub || notify != tc.notify {
				t.Fatalf("want %v, %v; got %v, %v", tc.pub, tc.notify, pub, notify)
			}
		})
	}
}
