package almanack

import (
	"flag"
	"net/http"

	"github.com/carlmjohnson/slackhook"
	"github.com/earthboundkid/flagx/v2"

	"github.com/spotlightpa/almanack/internal/anf"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/healthchecksio"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/jsonfeed"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func AddFlags(fl *flag.FlagSet) func() (svc Services, err error) {
	arcFeedURL := fl.String("src-feed", "", "source `URL` for Arc feed")
	mailchimpSignupURL := fl.String("mc-signup-url", "http://example.com", "`URL` to redirect users to for MailChimp signup")
	netlifyHookSecret := fl.String("netlify-webhook-secret", "", "`shared secret` to authorize Netlify identity webhook")
	newsfeed := jsonfeed.AddFlags(fl)
	anfService := anf.AddFlags(fl)

	pg, tx := db.AddFlags(fl, "postgres", "PostgreSQL database `URL`")
	slackSocial := slackhook.New(slackhook.MockClient)
	fl.Var(slackSocial, "slack-social-url", "Slack hook endpoint `URL` for social")
	slackTech := slackhook.New(slackhook.MockClient)
	fl.Var(slackTech, "slack-hook-url", "Slack tech channel endpoint `URL`")
	getS3Store := aws.AddFlags(fl)
	getGithub := github.AddFlags(fl)
	getIndex := index.AddFlags(fl)
	getNewsletter := mailchimp.AddFlags(fl)
	var gsvc google.Service
	gsvc.AddFlags(fl)
	mailServiceAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp v2")
	mailServiceListID := fl.String("mc-list-id", "", "List `ID` MailChimp v2 campaign")
	hc := fl.String("healthchecks-uuid", "", "`UUID` for Healthchecks.io alert")

	return func() (svc Services, err error) {
		if err = flagx.MustHave(fl, "postgres"); err != nil {
			return
		}

		client := http.Client{
			Transport: almlog.HTTPTransport,
		}

		is, fs := getS3Store()
		mc := mailchimp.NewMailService(*mailServiceAPIKey, *mailServiceListID, &client)
		anfService.Client = &client

		return Services{
			arcFeedURL:           *arcFeedURL,
			MailchimpSignupURL:   *mailchimpSignupURL,
			NetlifyWebhookSecret: *netlifyHookSecret,
			Client:               &client,
			Queries:              pg,
			Tx:                   tx,
			ContentStore:         getGithub(),
			SlackSocial:          slackSocial,
			SlackTech:            slackTech,
			ImageStore:           is,
			FileStore:            fs,
			Indexer:              getIndex(),
			NewletterService:     getNewsletter(&client),
			Gsvc:                 &gsvc,
			EmailService:         mc,
			HC:                   healthchecksio.New(*hc, &client),
			NewsFeed:             newsfeed,
			ANF:                  anfService,
		}, nil
	}
}
