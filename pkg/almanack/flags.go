package almanack

import (
	"flag"
	"net/http"

	"github.com/carlmjohnson/flagx"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/herokuapi"
	"github.com/spotlightpa/almanack/internal/httpcache"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/pkg/common"
)

func AddFlags(fl *flag.FlagSet) func() (svc Service, err error) {
	arcFeedURL := fl.String("src-feed", "", "source `URL` for Arc feed")
	mailchimpSignupURL := fl.String("mc-signup-url", "http://example.com", "`URL` to redirect users to for MailChimp signup")

	cache := fl.Bool("cache", false, "use in-memory cache for http requests")
	pg := db.AddFlags(fl, "postgres", "PostgreSQL database `URL`")
	slackURL := fl.String("slack-social-url", "", "Slack hook endpoint `URL` for social")
	heroku := herokuapi.AddFlags(fl)
	getS3Store := aws.AddFlags(fl)
	getGithub := github.AddFlags(fl)
	getIndex := index.AddFlags(fl)
	getNewsletter := mailchimp.AddFlags(fl)
	getGoogle := google.AddFlags(fl)
	mailServiceAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp v2")
	mailServiceListID := fl.String("mc-list-id", "", "List `ID` MailChimp v2 campaign")

	return func() (svc Service, err error) {
		// Get PostgreSQL URL from Heroku if possible, else get it from flag
		if err = heroku.Configure(map[string]string{
			"postgres":    "DATABASE_URL",
			"google-json": "ALMANACK_GOOGLE_JSON",
		}); err != nil {
			return
		}
		if err = flagx.MustHave(fl, "postgres"); err != nil {
			return
		}

		client := *http.DefaultClient
		if *cache {
			httpcache.SetRounderTripper(&client)
		}

		if svc.ContentStore, err = getGithub(); err != nil {
			common.Logger.Printf("could not connect to Github: %v", err)
			return
		}

		is, fs := getS3Store()
		mc := mailchimp.NewMailService(*mailServiceAPIKey, *mailServiceListID, &client)

		return Service{
			arcFeedURL:         *arcFeedURL,
			MailchimpSignupURL: *mailchimpSignupURL,
			Client:             &client,
			Queries:            pg,
			ContentStore:       svc.ContentStore,
			SlackClient:        slack.New(*slackURL),
			ImageStore:         is,
			FileStore:          fs,
			Indexer:            getIndex(),
			NewletterService:   getNewsletter(&client),
			gsvc:               getGoogle(),
			EmailService:       mc,
		}, nil
	}
}
