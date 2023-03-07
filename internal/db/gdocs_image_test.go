package db_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestGDocsImageUpsert(t *testing.T) {
	p := createTestDB(t)

	q := db.New(p)
	ctx := context.Background()
	pairs, err := json.Marshal([][]string{
		{"123", "http://example.com/image1.png"},
	})
	be.NilErr(t, err)

	err = q.UpsertGDocsIDObjectID(ctx, db.UpsertGDocsIDObjectIDParams{
		GDocsID:        "123",
		ObjectUrlPairs: pairs,
	})
	be.NilErr(t, err)

	pairs, err = json.Marshal([][]string{
		{"123", "http://example.com/image1.png"},
		{"456", "http://example.com/image2.png"},
		{"789", "http://example.com/image3.png"},
	})
	be.NilErr(t, err)

	err = q.UpsertGDocsIDObjectID(ctx, db.UpsertGDocsIDObjectIDParams{
		GDocsID:        "123",
		ObjectUrlPairs: pairs,
	})
	be.NilErr(t, err)

	rows, err := q.ListGDocsImagesWhereUnset(ctx)
	be.NilErr(t, err)
	be.True(t, len(rows) == 3)
}
