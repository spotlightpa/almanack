package integration_test

import (
	"net/http"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/requests"
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
	var created1 db.Promotion
	{ // Create first promotion
		var created db.Promotion
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(db.Promotion{
				Name:        "Banner Ad",
				Description: "A test banner",
				Link:        "https://example.com/",
				Width:       728,
				Height:      90,
				ImageUrls:       []string{"https://example.com/banner.png"},
			}).
			ToJSON(&created).
			Fetch(ctx))
		be.Nonzero(t, created.ID)
		be.Equal(t, "Banner Ad", created.Name)
		be.Equal(t, int32(728), created.Width)
		be.Equal(t, int32(90), created.Height)
		created1 = created
	}
	{ // Create second promotion
		var created db.Promotion
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(db.Promotion{
				Name:        "Sidebar Ad",
				Description: "A sidebar promotion",
				Link:        "https://example.org/",
				Width:       300,
				Height:      250,
				ImageUrls:       []string{"https://example.org/sidebar.png"},
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
			be.Equal(t, int32(300), p.Width)
		}
	}
	{ // Create with nil items — must not fail with NOT NULL violation
		var created db.Promotion
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(db.Promotion{
				Name:   "Nil items promo",
				Width:  300,
				Height: 250,
				// Items intentionally omitted (nil)
			}).
			ToJSON(&created).
			Fetch(ctx))
		be.Nonzero(t, created.ID)
		be.AllEqual(t, []string{}, created.ImageUrls)
	}
	{ // Update the first promotion
		var updated db.Promotion
		be.NilErr(t, rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(db.Promotion{
				ID:          created1.ID,
				Name:        "Banner Ad Updated",
				Description: created1.Description,
				Link:        created1.Link,
				Width:       created1.Width,
				Height:      created1.Height,
				ImageUrls:       created1.ImageUrls,
			}).
			ToJSON(&updated).
			Fetch(ctx))
		be.Equal(t, created1.ID, updated.ID)
		be.Equal(t, "Banner Ad Updated", updated.Name)
		be.Equal(t, int32(728), updated.Width)
	}
}

func TestDeletePromotionEndpoint(t *testing.T) {
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	ctx := t.Context()

	svc := almsvc.Services{
		DB:      dbhandle,
		Queries: dbhandle.Queries(),
		Auth:    netlifyid.MockAuthService{},
	}
	rb := newTestServer(t, svc)

	// Create a promotion to delete
	var created db.Promotion
	be.NilErr(t, rb.Clone().
		Path("/api/promotion").
		Method(http.MethodPost).
		BodyJSON(db.Promotion{
			Name:   "To Be Deleted",
			Link:   "https://example.com/",
			Width:  300,
			Height: 250,
			ImageUrls:  []string{"https://example.com/img.png"},
		}).
		ToJSON(&created).
		Fetch(ctx))
	be.Nonzero(t, created.ID)

	// Missing ID returns 400 Bad Request
	err := rb.Clone().
		Path("/api/promotion-delete").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"id": 1_000}).
		Fetch(ctx)
	re := new(requests.ResponseError)
	be.ErrorAs(t, &re, err)
	be.Equal(t, http.StatusBadRequest, re.StatusCode)

	// Delete the promotion
	be.NilErr(t, rb.Clone().
		Path("/api/promotion-delete").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"id": created.ID}).
		Fetch(ctx))

	// Confirm it no longer appears in the list
	var listResp struct {
		Promotions []db.Promotion `json:"promotions"`
	}
	be.NilErr(t, rb.Clone().
		Path("/api/promotion").
		ToJSON(&listResp).
		Fetch(ctx))
	for _, p := range listResp.Promotions {
		be.Unequal(t, created.ID, p.ID)
	}
}
