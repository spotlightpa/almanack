package integration_test

import (
	"fmt"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/utils/pgxutil"
	"testing"
)

func TestMap(t *testing.T) {
	be := assert.FailNow(t)
	dbhandle := createTestDB(t)
	q := dbhandle.Queries()

	ctx := t.Context()
	const testpath = "test/hello.md"
	_, err := q.CreatePage(ctx, db.CreatePageParams{
		FilePath:   testpath,
		SourceType: "testing",
	})
	be.Zero(err)
	// create again
	_, err = q.CreatePage(ctx, db.CreatePageParams{
		FilePath:   testpath,
		SourceType: "testing",
	})
	be.NotZero(err)
	p1, err := q.GetPageByFilePath(ctx, testpath)
	be.Zero(err)
	be.Equal(p1.FilePath, testpath)
	p2, err := q.UpdatePage(ctx, db.UpdatePageParams{
		ID:             p1.ID,
		SetFrontmatter: true,
		Frontmatter: db.Map{
			"hello":  "world",
			"bool":   true,
			"number": 1,
		},
		SetBody:     true,
		Body:        "hello",
		ScheduleFor: pgxutil.NullTime,
	})
	be.Zero(err)
	be.Equal(p2.FilePath, testpath)
	be.Equal(p2.Body, "hello")
	be.Equal(fmt.Sprint(p2.Frontmatter), "map[bool:true hello:world number:1]")
	p3, err := q.GetPageByFilePath(ctx, testpath)
	be.Zero(err)
	be.Equal(p3.FilePath, testpath)
	be.Equal(p3.Body, "hello")
	be.Equal(fmt.Sprint(p3.Frontmatter), "map[bool:true hello:world number:1]")
}
