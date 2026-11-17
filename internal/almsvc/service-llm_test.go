package almsvc_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/carlmjohnson/requests/reqtest"
	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/services/anthropic"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"gocloud.dev/blob/driver"
)

func TestDescribeImage(t *testing.T) {
	be := assert.FailsNow(t)
	almlog.UseTestLogger(t)
	t.Cleanup(aws.UseMockSigner(func(ctx context.Context, key string, opts *driver.SignedURLOptions) (string, error) {
		return "http://example.com/signed?key=" + key, nil
	}))

	var data any
	svc := almsvc.Services{
		ImageStore: aws.NewBlobStore("mem://"),
		LLM: &anthropic.Service{
			Client: &http.Client{
				Transport: reqtest.ReplayJSON(200, &data),
			}}}

	err := be.NotOK(svc.DescribeImage(t.Context(), "abc/123"))
	be.Equal(resperr.StatusCode(err), 502)
	be.Match(err.Error(), `no text returned by LLM API`)

	data = anthropic.CreateMessageResponse{
		Content: []anthropic.Content{
			{Type: "text", Text: "hello"},
		}}
	desc := be.OK(svc.DescribeImage(t.Context(), "abc/123"))
	be.Equal(desc, "hello")
}
