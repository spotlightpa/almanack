package integration_test

import (
	"fmt"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/utils/pgxutil"
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
	p1 := be.OK(q.GetPageByFilePath(ctx, testpath))
	be.Equal(p1.FilePath, testpath)
	p2 := be.OK(q.UpdatePage(ctx, db.UpdatePageParams{
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
	}))
	be.
		Equal(p2.FilePath, testpath).
		Equal(p2.Body, "hello").
		Equal(fmt.Sprint(p2.Frontmatter), "map[bool:true hello:world number:1]")
	p3 := be.OK(q.GetPageByFilePath(ctx, testpath))
	be.
		Equal(p3.FilePath, testpath).
		Equal(p3.Body, "hello").
		Equal(fmt.Sprint(p3.Frontmatter), "map[bool:true hello:world number:1]")
}
