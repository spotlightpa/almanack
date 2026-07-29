package integration_test

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/requests/reqtest"
	"github.com/spotlightpa/almanack/internal/almapp"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
)

func TestAuthorizedAddressesEndpoints(t *testing.T) {
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	ctx := t.Context()

	srv := httptest.NewServer(almapp.NewHandler(almsvc.Services{
		DB:      dbhandle,
		Queries: dbhandle.Queries(),
		Auth:    netlifyid.MockAuthService{},
	}))
	t.Cleanup(srv.Close)

	base := requests.New(reqtest.Server(srv)).Header("Authorization", "Bearer mock")

	var addrResp struct {
		Addresses []string `json:"addresses"`
	}

	// Add two addresses via POST.
	be.NilErr(t, base.Clone().
		Path("/api/authorized-addresses").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"address": "alice@example.com"}).
		ToJSON(&addrResp).
		Fetch(ctx))
	be.True(t, slices.Contains(addrResp.Addresses, "alice@example.com"))

	be.NilErr(t, base.Clone().
		Path("/api/authorized-addresses").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"address": "bob@example.com"}).
		ToJSON(&addrResp).
		Fetch(ctx))
	be.True(t, slices.Contains(addrResp.Addresses, "alice@example.com"))
	be.True(t, slices.Contains(addrResp.Addresses, "bob@example.com"))

	// GET the list and confirm both addresses appear.
	var list struct {
		Addresses []string `json:"addresses"`
	}
	be.NilErr(t, base.Clone().
		Path("/api/authorized-addresses").
		ToJSON(&list).
		Fetch(ctx))
	be.AtLeastLength(t, 2, list.Addresses)
	be.True(t, slices.Contains(list.Addresses, "alice@example.com"))
	be.True(t, slices.Contains(list.Addresses, "bob@example.com"))
}
