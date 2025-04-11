package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestMap(t *testing.T) {
	p := createTestDB(t)
	q := db.New(p)

	ctx := context.Background()
	const testpath = "test/hello.md"
	_, err := q.CreatePageV2(ctx, db.CreatePageV2Params{
		FilePath:   testpath,
		SourceType: "testing",
	})
	be.NilErr(t, err)
	// create again
	_, err = q.CreatePageV2(ctx, db.CreatePageV2Params{
		FilePath:   testpath,
		SourceType: "testing",
	})
	be.Nonzero(t, err)
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
