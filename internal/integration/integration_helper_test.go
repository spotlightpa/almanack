package integration_test

import (
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spotlightpa/almanack/internal/almapp"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
)

var (
	once    sync.Once
	pool    *pgxpool.Pool
	poolErr error
)

func createTestDB(t *testing.T) *db.Handle {
	t.Helper()
	dbURL := os.Getenv("ALMANACK_POSTGRES")
	if dbURL == "" {
		t.Skip("ALMANACK_POSTGRES not set")
	}
	once.Do(func() {
		pool, poolErr = db.CreateTestDatabase(dbURL)
	})
	be.NilErr(t, poolErr)
	return db.NewHandle(pool)
}

// newTestServer starts an httptest.Server backed by the real router and a
// mock auth service. It registers cleanup and returns a *requests.Builder
// pre-configured with the server's base URL and a mock Authorization header.
func newTestServer(t *testing.T, svc almsvc.Services) *requests.Builder {
	t.Helper()
	srv := httptest.NewServer(almapp.NewHandler(svc))
	t.Cleanup(srv.Close)
	return requests.New(reqtest.Server(srv)).Header("Authorization", "Bearer mock")
}
