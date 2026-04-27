package almservices

import (
	"flag"
	"net/http"

	"github.com/earthboundkid/flagx/v2"
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
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func AddFlags(fl *flag.FlagSet) func() (svc Services, err error) {
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
	yt := youtube.AddFlags(fl)

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
			YT:                   yt,
		}, nil
	}
}
