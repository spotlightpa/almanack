// Package healthchecksio is a wrapper for API calls to HealthChecks.io
package healthchecksio

import (
	"context"
	"net/http"
	"strings"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
)

// Client is a convenient way to ping HealthChecks.io
type Client struct {
	rb *requests.Builder
}

// New returns a configured client. If c is nil, http.DefaultClient is used.
func New(uuid string, c *http.Client) Client {
	if !strings.HasSuffix(uuid, "/") {
		uuid = uuid + "/"
	}
	return Client{
		requests.
			URL("https://hc-ping.com").
			Path(uuid).
			UserAgent("spotlightpa/almanack").
			Client(c),
	}
}

// Start calls the start HealthChecks.io endpoint
func (cl Client) Start(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	return cl.rb.Clone().
		Path("start").
		Fetch(ctx)
}

// Status calls the HealthChecks.io status endpoint
func (cl Client) Status(ctx context.Context, code uint8, msg []byte) (err error) {
	defer errorx.Trace(&err)

	return cl.rb.Clone().
		Pathf("%d", code).
		BodyBytes(msg).
		Fetch(ctx)
}

func (cl Client) Success(ctx context.Context, msg []byte) (err error) {
	defer errorx.Trace(&err)

	return cl.Status(ctx, 0, msg)
}

func (cl Client) Fail(ctx context.Context, msg []byte) (err error) {
	defer errorx.Trace(&err)

	return cl.rb.Clone().
		Path("fail").
		BodyBytes(msg).
		Fetch(ctx)
}
