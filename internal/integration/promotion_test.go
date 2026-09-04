package integration_test

import (
	"github.com/carlmjohnson/requests"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
	"net/http"
	"testing"
)

func TestPromotionEndpoints(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	rb := newTestServer(t, almsvc.Services{
		DB:      dbhandle,
		Queries: dbhandle.Queries(),
		Auth:    netlifyid.MockAuthService{},
	})
	ctx := t.Context()

	var created1 db.Promotion
	{ // Create first promotion
		var created db.Promotion
		be.Zero(rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(db.Promotion{
				Name:             "Banner Ad",
				Description:      "A test banner",
				Link:             "https://example.com/",
				Width:            728,
				Height:           90,
				ImageUrls:        []string{"https://example.com/banner.png"},
				ImageDescription: "A colorful banner",
				BannerLabel:      "Sponsored by Acme",
				BannerLabelLink:  "https://acme.example.com/",
			}).
			ToJSON(&created).
			Fetch(ctx))
		be.NotZero(created.ID)
		be.Equal(created.Name, "Banner Ad")
		be.Equal(created.Width, int32(728))
		be.Equal(created.Height, int32(90))
		be.Equal(created.ImageDescription, "A colorful banner")
		be.Equal(created.BannerLabel, "Sponsored by Acme")
		be.Equal(created.BannerLabelLink, "https://acme.example.com/")
		created1 = created
	}
	{ // Create second promotion
		var created db.Promotion
		be.Zero(rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(db.Promotion{
				Name:        "Sidebar Ad",
				Description: "A sidebar promotion",
				Link:        "https://example.org/",
				Width:       300,
				Height:      250,
				ImageUrls:   []string{"https://example.org/sidebar.png"},
			}).
			ToJSON(&created).
			Fetch(ctx))
		be.NotZero(created.ID)
		be.Equal(created.Name, "Sidebar Ad")
	}
	{ // List all promotions (no text filter) — NextPage must be absent when results fit in one page
		var listResp struct {
			Promotions []db.Promotion `json:"promotions"`
			NextPage   string         `json:"next_page"`
		}
		be.Zero(rb.Clone().
			Path("/api/promotion").
			ToJSON(&listResp).
			Fetch(ctx))
		be.AtLeastLength(listResp.Promotions, 2)
		be.Equal(listResp.NextPage, "")
	}
	{ // List promotions with FTS text filter
		var ftsResp struct {
			Promotions []db.Promotion `json:"promotions"`
		}
		be.Zero(rb.Clone().
			Path("/api/promotion").
			Param("text", "sidebar").
			ToJSON(&ftsResp).
			Fetch(ctx))
		be.AtLeastLength(ftsResp.Promotions, 1)
		be.Equal(ftsResp.Promotions[0].Name, "Sidebar Ad")
	}
	{ // List promotions filtered by width
		var widthResp struct {
			Promotions []db.Promotion `json:"promotions"`
		}
		be.Zero(rb.Clone().
			Path("/api/promotion").
			Param("width", "300").
			ToJSON(&widthResp).
			Fetch(ctx))
		be.AtLeastLength(widthResp.Promotions, 1)
		for _, p := range widthResp.Promotions {
			be.Equal(p.Width, int32(300))
		}
	}
	{ // Create with nil image_urls — must not fail with NOT NULL violation
		var created db.Promotion
		be.Zero(rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(db.Promotion{
				Name:   "Nil items promo",
				Width:  300,
				Height: 250,
				// ImageUrls intentionally omitted (nil)
			}).
			ToJSON(&created).
			Fetch(ctx))
		be.NotZero(created.ID)
		be.SlicesEqual(created.ImageUrls, []string{})
	}
	{ // Update the first promotion
		var updated db.Promotion
		be.Zero(rb.Clone().
			Path("/api/promotion").
			Method(http.MethodPost).
			BodyJSON(db.Promotion{
				ID:               created1.ID,
				Name:             "Banner Ad Updated",
				Description:      created1.Description,
				Link:             created1.Link,
				Width:            created1.Width,
				Height:           created1.Height,
				ImageUrls:        created1.ImageUrls,
				ImageDescription: "Updated alt text",
				BannerLabel:      created1.BannerLabel,
				BannerLabelLink:  created1.BannerLabelLink,
			}).
			ToJSON(&updated).
			Fetch(ctx))
		be.Equal(updated.ID, created1.ID)
		be.Equal(updated.Name, "Banner Ad Updated")
		be.Equal(updated.Width, int32(728))
		be.Equal(updated.ImageDescription, "Updated alt text")
	}
}

func TestDeletePromotionEndpoint(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	rb := newTestServer(t, almsvc.Services{
		DB:      dbhandle,
		Queries: dbhandle.Queries(),
		Auth:    netlifyid.MockAuthService{},
	})
	ctx := t.Context()

	// Create a promotion to delete
	var created db.Promotion
	be.Zero(rb.Clone().
		Path("/api/promotion").
		Method(http.MethodPost).
		BodyJSON(db.Promotion{
			Name:      "To Be Deleted",
			Link:      "https://example.com/",
			Width:     300,
			Height:    250,
			ImageUrls: []string{"https://example.com/img.png"},
		}).
		ToJSON(&created).
		Fetch(ctx))
	be.NotZero(created.ID)

	// Nonexistent ID returns 400 Bad Request
	err := rb.Clone().
		Path("/api/promotion-delete").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"id": 1_000}).
		Fetch(ctx)
	re := be.ErrorAsType[*requests.ResponseError](err)
	be.Equal(re.StatusCode, http.StatusBadRequest)

	// Delete the promotion
	be.Zero(rb.Clone().
		Path("/api/promotion-delete").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"id": created.ID}).
		Fetch(ctx))

	// Confirm it no longer appears in the list
	var listResp struct {
		Promotions []db.Promotion `json:"promotions"`
	}
	be.Zero(rb.Clone().
		Path("/api/promotion").
		ToJSON(&listResp).
		Fetch(ctx))
	for _, p := range listResp.Promotions {
		be.NotEqual(p.ID, created.ID)
	}
}
