package almanack

import (
	"net/http"

	"github.com/carlmjohnson/slackhook"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/healthchecksio"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/plausible"
)

type Services struct {
	arcFeedURL           string
	MailchimpSignupURL   string
	NetlifyWebhookSecret string
	Client               *http.Client
	Queries              *db.Queries
	Tx                   *db.Txable
	github.ContentStore
	ImageStore       aws.BlobStore
	FileStore        aws.BlobStore
	SlackSocial      *slackhook.Client
	SlackTech        *slackhook.Client
	Indexer          index.Indexer
	NewletterService mailchimp.V3
	Gsvc             *google.Service
	mailchimp.EmailService
	Plausible plausible.API
	HC        healthchecksio.Client
}
