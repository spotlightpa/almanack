package integration_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
)

func TestAuthorizedAddressesEndpoints(t *testing.T) {
	be := assert.FailNow(t)
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
	be.Zero(rb.Clone().
		Path("/api/authorized-addresses").
		ToJSON(&list).
		Fetch(ctx))
	be.Zero(list.Addresses)

	// Add two addresses via POST.
	be.Zero(rb.Clone().
		Path("/api/authorized-addresses").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"address": "alice@example.com"}).
		ToJSON(&list).
		Fetch(ctx))
	be.True(slices.Contains(list.Addresses, "alice@example.com"))

	be.Zero(rb.Clone().
		Path("/api/authorized-addresses").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"address": "bob@example.com"}).
		ToJSON(&list).
		Fetch(ctx))
	be.
		True(slices.Contains(list.Addresses, "alice@example.com")).
		True(slices.Contains(list.Addresses, "bob@example.com"))

	// GET the list and confirm both addresses appear.
	be.Zero(rb.Clone().
		Path("/api/authorized-addresses").
		ToJSON(&list).
		Fetch(ctx))
	be.
		AtLeastLength(list.Addresses, 2).
		True(slices.Contains(list.Addresses, "alice@example.com")).
		True(slices.Contains(list.Addresses, "bob@example.com"))
}
