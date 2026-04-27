package almservices

import (
	"net/http"

	"github.com/earthboundkid/slackhook/v2"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/anf"
	"github.com/spotlightpa/almanack/internal/services/aws"
	"github.com/spotlightpa/almanack/internal/services/github"
	"github.com/spotlightpa/almanack/internal/services/google"
	"github.com/spotlightpa/almanack/internal/services/healthchecksio"
	"github.com/spotlightpa/almanack/internal/services/index"
	"github.com/spotlightpa/almanack/internal/services/jsonfeed"
	"github.com/spotlightpa/almanack/internal/services/mailchimp"
	"github.com/spotlightpa/almanack/internal/services/youtube"
)

type Services struct {
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
	NewsFeed         *jsonfeed.NewsFeed
	ANF              *anf.Service
	mailchimp.EmailService
	HC healthchecksio.Client
	YT *youtube.Feed
}
