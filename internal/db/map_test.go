package db_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spotlightpa/almanack/internal/db"
)

func TestMap(t *testing.T) {
	dbURL := os.Getenv("ALMANACK_TEST_DATABASE")
	if dbURL == "" {
		t.Skip("ALMANACK_TEST_DATABASE not set")
	}
	q, err := db.Open(dbURL)
	check(t, err, "could not open DB")
	ctx := context.Background()
	const testpath = "test/hello.md"
	err = q.EnsurePage(ctx, testpath)
	check(t, err, "could not create page")
	// create again
	err = q.EnsurePage(ctx, testpath)
	check(t, err, "creation not idempotent")
	p1, err := q.GetPage(ctx, testpath)
	check(t, err, "could not get page")
	eq(t, testpath, p1.Path)
	p2, err := q.UpdatePage(ctx, db.UpdatePageParams{
		SetFrontmatter: true,
		Frontmatter: db.Map{
			"hello":  "world",
			"bool":   true,
			"number": 1,
		},
		SetBody: true,
		Body:    "hello",
		Path:    testpath,
	})
	check(t, err, "could not update page")
	eq(t, testpath, p2.Path)
	eq(t, "hello", p2.Body)
	eq(t, "map[bool:true hello:world number:1]", fmt.Sprint(p2.Frontmatter))
	p3, err := q.GetPage(ctx, testpath)
	check(t, err, "could not get page")
	eq(t, testpath, p3.Path)
	eq(t, "hello", p3.Body)
	eq(t, "map[bool:true hello:world number:1]", fmt.Sprint(p3.Frontmatter))
}
