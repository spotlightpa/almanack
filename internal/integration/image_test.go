package integration_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
	"gocloud.dev/blob/driver"
)

func TestCreateSignedUploadStoresDimensions(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	t.Cleanup(aws.UseMockSigner(func(ctx context.Context, key string, opts *driver.SignedURLOptions) (string, error) {
		return "http://example.com/signed?key=" + key, nil
	}))
	rb := newTestServer(t, almsvc.Services{
		DB:         dbhandle,
		Queries:    dbhandle.Queries(),
		Auth:       netlifyid.MockAuthService{},
		ImageStore: aws.NewBlobStore("mem://"),
	})
	ctx := t.Context()

	var resp struct {
		SignedURL string `json:"signed-url"`
		Filename  string `json:"filename"`
	}

	// Upload with explicit width and height
	be.NilError(rb.Clone().
		Path("/api/create-signed-upload").
		Method(http.MethodPost).
		BodyJSON(map[string]any{
			"type":   "image/jpeg",
			"width":  1920,
			"height": 1080,
		}).
		ToJSON(&resp).
		Fetch(ctx))
	be.Truthy(resp.Filename)
	be.Truthy(resp.SignedURL)

	// Verify the image row was created with the correct dimensions
	img := be.OK(dbhandle.Queries().GetImageByPath(ctx, resp.Filename))
	be.
		Equal(img.Width, int32(1920)).
		Equal(img.Height, int32(1080))

	// Upload without width/height — dimensions should default to 0
	var resp2 struct {
		Filename string `json:"filename"`
	}
	be.NilError(rb.Clone().
		Path("/api/create-signed-upload").
		Method(http.MethodPost).
		BodyJSON(map[string]any{
			"type": "image/png",
		}).
		ToJSON(&resp2).
		Fetch(ctx))
	be.Truthy(resp2.Filename)

	img2 := be.OK(dbhandle.Queries().GetImageByPath(ctx, resp2.Filename))
	be.
		Equal(img2.Width, int32(0)).
		Equal(img2.Height, int32(0))

	// Confirm an unsupported type returns an error
	err := rb.Clone().
		Path("/api/create-signed-upload").
		Method(http.MethodPost).
		BodyJSON(map[string]any{"type": "image/gif"}).
		Fetch(ctx)
	be.Truthy(err)
}

func TestCreateSignedUploadUpsertPreservesExistingDimensions(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	ctx := t.Context()

	// Pre-create an image row with known dimensions via direct DB upsert
	existing := be.OK(dbhandle.Queries().UpsertImage(ctx, db.UpsertImageParams{
		Path:   "2024/01/test-existing.jpeg",
		Type:   "jpeg",
		Width:  800,
		Height: 600,
	}))
	be.
		Equal(existing.Width, int32(800)).
		Equal(existing.Height, int32(600))

	// Upsert again with 0 dimensions — existing values must be preserved
	updated := be.OK(dbhandle.Queries().UpsertImage(ctx, db.UpsertImageParams{
		Path:  "2024/01/test-existing.jpeg",
		Type:  "jpeg",
		Width: 0,
		// Height 0 as well (zero value)
	}))
	be.
		Equal(updated.Width, int32(800)).
		Equal(updated.Height, int32(600))
}
