package integration_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/services/anthropic"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
	"gocloud.dev/blob/driver"
)

func TestDescribeImage(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	t.Cleanup(aws.UseMockSigner(func(ctx context.Context, key string, opts *driver.SignedURLOptions) (string, error) {
		return "http://example.com/signed?key=" + key, nil
	}))

	data := anthropic.CreateMessageResponse{
		Content: []anthropic.Content{
			{Type: "text", Text: "hello"},
		}}
	rb := newTestServer(t, almsvc.Services{
		Auth:       netlifyid.MockAuthService{},
		ImageStore: aws.NewBlobStore("mem://"),
		LLM: &anthropic.Service{
			Client: &http.Client{
				Transport: reqtest.ReplayJSON(200, data),
			}}})

	var desc string
	be.NilError(rb.Clone().
		Path("/api/describe-image").
		Param("image", "abc/123").
		ToJSON(&desc).
		Fetch(t.Context()))
	be.Equal(desc, "hello")
}
