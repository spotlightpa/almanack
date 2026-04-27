package mailchimp

import (
	"flag"
	"net/http"
	"strings"

	"github.com/carlmjohnson/requests"
)

func AddFlags(fs *flag.FlagSet) func(c *http.Client) V3 {
	apiKey := fs.String("mcnl-api-key", "", "`API key` for MailChimp newsletter archive")
	listID := fs.String("mcnl-list-id", "", "List `ID` for MailChimp newsletter archive")

	return func(c *http.Client) V3 {
		return NewV3(*apiKey, *listID, c)
	}
}

type V3 struct {
	apiKey string
	listID string
	cl     *http.Client
}

func NewV3(apiKey, listID string, cl *http.Client) V3 {
	return V3{apiKey, listID, cl}
}

func (v3 V3) config(rb *requests.Builder) {
	// API keys end with 123XYZ-us1, where us1 is the datacenter
	_, datacenter, _ := strings.Cut(v3.apiKey, "-")
	rb.
		BaseURL("https://dc.api.mailchimp.com/3.0/").
		Client(v3.cl).
		BasicAuth("", v3.apiKey).
		Hostf("%s.api.mailchimp.com", datacenter)
}
