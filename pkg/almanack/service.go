package almanack

import (
	"net/http"

	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/slack"
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
	SlackSocial      slack.Client
	SlackTech        slack.Client
	Indexer          index.Indexer
	NewletterService mailchimp.V3
	gsvc             *google.Service
	mailchimp.EmailService
}
