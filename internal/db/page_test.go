package db_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

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
				URLPath: sql.NullString{
					Valid:  true,
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
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.Page.SetURLPath()
			if tc.Page.URLPath.Valid != (tc.Page.URLPath.String != "") {
				t.Fatalf("bad validity")
			}

			if tc.Page.URLPath.String != tc.string {
				t.Fatalf("got %v", tc.Page.URLPath)
			}
		})
	}
}
