package integration_test

import (
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
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

// createTestDB checks ALMANACK_POSTGRES and
// creates a database called almanack_test at the provided address.
func createTestDB(t *testing.T) *db.Handle {
	t.Helper()
	dbURL := os.Getenv("ALMANACK_POSTGRES")
	if dbURL == "" {
		t.Skip("ALMANACK_POSTGRES not set")
	}
	once.Do(func() {
		pool, poolErr = db.CreateTestDatabase(dbURL)
	})
	assert.FailNow(t).Zero(poolErr)
	return db.NewHandle(pool)
}

// newTestServer starts an httptest.Server backed by svc.
// It registers cleanup and returns a *requests.Builder
// pre-configured with the server's base URL and a mock Authorization header.
func newTestServer(t *testing.T, svc almsvc.Services) *requests.Builder {
	t.Helper()
	return requests.
		New(reqtest.Server(httptest.NewTestServer(t, almapp.NewHandler(svc)))).
		Header("Authorization", "Bearer mock")
}
