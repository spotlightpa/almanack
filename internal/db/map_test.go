package db_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestMap(t *testing.T) {
	dbURL := os.Getenv("ALMANACK_TEST_DATABASE")
	if dbURL == "" {
		t.Skip("ALMANACK_TEST_DATABASE not set")
	}
	p, err := db.Open(dbURL)
	q := db.New(p)
	be.NilErr(t, err)
	ctx := context.Background()
	const testpath = "test/hello.md"
	err = q.CreatePage(ctx, db.CreatePageParams{
		FilePath:   testpath,
		SourceType: "testing",
	})
	be.NilErr(t, err)
	// create again
	err = q.CreatePage(ctx, db.CreatePageParams{
		FilePath:   testpath,
		SourceType: "testing",
	})
	be.NilErr(t, err)
	p1, err := q.GetPageByFilePath(ctx, testpath)
	be.NilErr(t, err)
	be.Equal(t, testpath, p1.FilePath)
	p2, err := q.UpdatePage(ctx, db.UpdatePageParams{
		SetFrontmatter: true,
		Frontmatter: db.Map{
			"hello":  "world",
			"bool":   true,
			"number": 1,
		},
		SetBody:     true,
		Body:        "hello",
		FilePath:    testpath,
		ScheduleFor: db.NullTime,
	})
	be.NilErr(t, err)
	be.Equal(t, testpath, p2.FilePath)
	be.Equal(t, "hello", p2.Body)
	be.Equal(t, "map[bool:true hello:world number:1]", fmt.Sprint(p2.Frontmatter))
	p3, err := q.GetPageByFilePath(ctx, testpath)
	be.NilErr(t, err)
	be.Equal(t, testpath, p3.FilePath)
	be.Equal(t, "hello", p3.Body)
	be.Equal(t, "map[bool:true hello:world number:1]", fmt.Sprint(p3.Frontmatter))
}
