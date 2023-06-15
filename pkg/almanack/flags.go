package almanack

import (
	"flag"
	"net/http"

	"github.com/carlmjohnson/flagx"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/slack"
)

func AddFlags(fl *flag.FlagSet) func() (svc Services, err error) {
	arcFeedURL := fl.String("src-feed", "", "source `URL` for Arc feed")
	mailchimpSignupURL := fl.String("mc-signup-url", "http://example.com", "`URL` to redirect users to for MailChimp signup")
	netlifyHookSecret := fl.String("netlify-webhook-secret", "", "`shared secret` to authorize Netlify identity webhook")

	pg, tx := db.AddFlags(fl, "postgres", "PostgreSQL database `URL`")
	slackSocialURL := fl.String("slack-social-url", "", "Slack hook endpoint `URL` for social")
	slackTechURL := fl.String("slack-hook-url", "", "Slack tech channel endpoint `URL`")
	getS3Store := aws.AddFlags(fl)
	getGithub := github.AddFlags(fl)
	getIndex := index.AddFlags(fl)
	getNewsletter := mailchimp.AddFlags(fl)
	var gsvc google.Service
	gsvc.AddFlags(fl)
	mailServiceAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp v2")
	mailServiceListID := fl.String("mc-list-id", "", "List `ID` MailChimp v2 campaign")

	return func() (svc Services, err error) {
		if err = flagx.MustHave(fl, "postgres"); err != nil {
			return
		}

		client := *http.DefaultClient

		is, fs := getS3Store()
		mc := mailchimp.NewMailService(*mailServiceAPIKey, *mailServiceListID, &client)

		return Services{
			arcFeedURL:           *arcFeedURL,
			MailchimpSignupURL:   *mailchimpSignupURL,
			NetlifyWebhookSecret: *netlifyHookSecret,
			Client:               &client,
			Queries:              pg,
			Tx:                   tx,
			ContentStore:         getGithub(),
			SlackSocial:          slack.New(*slackSocialURL),
			SlackTech:            slack.New(*slackTechURL),
			ImageStore:           is,
			FileStore:            fs,
			Indexer:              getIndex(),
			NewletterService:     getNewsletter(&client),
			Gsvc:                 &gsvc,
			EmailService:         mc,
		}, nil
	}
}
