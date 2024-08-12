// Package healthchecksio is a wrapper for API calls to HealthChecks.io
package healthchecksio

import (
	"context"
	"net/http"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
)

// Client is a convenient way to ping HealthChecks.io
type Client struct {
	rb *requests.Builder
}

// New returns a configured client. If c is nil, http.DefaultClient is used.
func New(uuid string, c *http.Client) Client {
	return Client{
		requests.
			URL("https://hc-ping.com").
			Path(uuid).
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
func (cl Client) Status(ctx context.Context, code int, msg []byte) (err error) {
	defer errorx.Trace(&err)

	return cl.rb.Clone().
		Pathf("%d", code).
		BodyBytes(msg).
		Fetch(ctx)
}
