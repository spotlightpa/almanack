package integration_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/almlog"
)

func TestAuthorizedAddressesEndpoints(t *testing.T) {
	almlog.UseTestLogger(t)
	ctx := t.Context()
	base := newTestServer(t)

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
