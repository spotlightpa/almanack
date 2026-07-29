package integration_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
)

func TestAuthorizedAddressesEndpoints(t *testing.T) {
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	rb := newTestServer(t, almsvc.Services{
		DB:      dbhandle,
		Queries: dbhandle.Queries(),
		Auth:    netlifyid.MockAuthService{},
	})
	ctx := t.Context()

	var list struct {
		Addresses []string `json:"addresses"`
	}

	// Initially empty
	be.NilErr(t, rb.Clone().
		Path("/api/authorized-addresses").
		ToJSON(&list).
		Fetch(ctx))
	be.Zero(t, list.Addresses)

	// Add two addresses via POST.
	be.NilErr(t, rb.Clone().
		Path("/api/authorized-addresses").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"address": "alice@example.com"}).
		ToJSON(&list).
		Fetch(ctx))
	be.True(t, slices.Contains(list.Addresses, "alice@example.com"))

	be.NilErr(t, rb.Clone().
		Path("/api/authorized-addresses").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"address": "bob@example.com"}).
		ToJSON(&list).
		Fetch(ctx))
	be.True(t, slices.Contains(list.Addresses, "alice@example.com"))
	be.True(t, slices.Contains(list.Addresses, "bob@example.com"))

	// GET the list and confirm both addresses appear.
	be.NilErr(t, rb.Clone().
		Path("/api/authorized-addresses").
		ToJSON(&list).
		Fetch(ctx))
	be.AtLeastLength(t, 2, list.Addresses)
	be.True(t, slices.Contains(list.Addresses, "alice@example.com"))
	be.True(t, slices.Contains(list.Addresses, "bob@example.com"))
}
