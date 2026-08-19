package integration_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
)

func TestPromotionEndpoints(t *testing.T) {
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	ctx := t.Context()

	svc := almsvc.Services{
		DB:      dbhandle,
		Queries: dbhandle.Queries(),
		Auth:    netlifyid.MockAuthService{},
	}
	rb := newTestServer(t, svc)
	type promoReq struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Data        any    `json:"data"`
	}
	type promoData struct {
		Width  int64 `json:"width"`
		Height int64 `json:"height"`
		Items  []any `json:"items"`
	}

	var created1 db.Promotion
	{ // Create first promotion
		var created db.Promotion
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(promoReq{
				Name:        "Banner Ad",
				Description: "A test banner",
				Data:        promoData{Width: 728, Height: 90, Items: []any{}},
			}).
			ToJSON(&created).
			Fetch(ctx))
		be.Nonzero(t, created.ID)
		be.Equal(t, "Banner Ad", created.Name)
		be.Equal(t, int64(728), created.Width)
		be.Equal(t, int64(90), created.Height)
		created1 = created
	}
	{ // Create second promotion
		var created db.Promotion
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(promoReq{
				Name:        "Sidebar Ad",
				Description: "A sidebar promotion",
				Data:        promoData{Width: 300, Height: 250, Items: []any{}},
			}).
			ToJSON(&created).
			Fetch(ctx))
		be.Nonzero(t, created.ID)
		be.Equal(t, "Sidebar Ad", created.Name)
	}
	{ // List all promotions (no text filter)
		var listResp struct {
			Promotions []db.Promotion `json:"promotions"`
			NextPage   int32          `json:"next_page,string,omitempty"`
		}
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			ToJSON(&listResp).
			Fetch(ctx))
		be.AtLeastLength(t, 2, listResp.Promotions)
	}
	{ // List promotions with FTS text filter
		var ftsResp struct {
			Promotions []db.Promotion `json:"promotions"`
		}
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Param("text", "sidebar").
			ToJSON(&ftsResp).
			Fetch(ctx))
		be.AtLeastLength(t, 1, ftsResp.Promotions)
		be.Equal(t, "Sidebar Ad", ftsResp.Promotions[0].Name)
	}
	{ // List promotions filtered by width
		var widthResp struct {
			Promotions []db.Promotion `json:"promotions"`
		}
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Param("width", "300").
			ToJSON(&widthResp).
			Fetch(ctx))
		be.AtLeastLength(t, 1, widthResp.Promotions)
		for _, p := range widthResp.Promotions {
			be.Equal(t, int64(300), p.Width)
		}
	}
	{ // Update the first promotion
		var updated db.Promotion
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(promoReq{
				ID:          created1.ID,
				Name:        "Banner Ad Updated",
				Description: created1.Description,
				Data:        promoData{Width: created1.Width, Height: created1.Height, Items: []any{}},
			}).
			ToJSON(&updated).
			Fetch(ctx))
		be.Equal(t, created1.ID, updated.ID)
		be.Equal(t, "Banner Ad Updated", updated.Name)
		be.Equal(t, int64(728), updated.Width)
	}
}
