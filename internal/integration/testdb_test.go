package integration_test

import (
	"os"
	"sync"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spotlightpa/almanack/internal/db"
)

var (
	once    sync.Once
	pool    *pgxpool.Pool
	poolErr error
)

func createTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dbURL := os.Getenv("ALMANACK_POSTGRES")
	if dbURL == "" {
		t.Skip("ALMANACK_POSTGRES not set")
	}
	once.Do(func() {
		pool, poolErr = db.CreateTestDatabase(dbURL)
	})
	be.NilErr(t, poolErr)
	return pool
}
