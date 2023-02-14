package db_test

import (
	"os"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spotlightpa/almanack/internal/db"
)

func createTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("ALMANACK_POSTGRES")
	if dbURL == "" {
		t.Skip("ALMANACK_POSTGRES not set")
	}
	p, err := db.CreateTestDatabase(dbURL)
	be.NilErr(t, err)
	return p
}
