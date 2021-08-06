package almanack

import (
	"flag"
	"net/http"

	"github.com/carlmjohnson/flagext"
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

func Flags(fl *flag.FlagSet) func(common.Logger) (svc Service, err error) {
	cache := fl.Bool("cache", false, "use in-memory cache for http requests")
	pg := db.FlagVar(fl, "postgres", "PostgreSQL database `URL`")
	slackURL := fl.String("slack-social-url", "", "Slack hook endpoint `URL` for social")
	heroku := herokuapi.ConfigureFlagSet(fl)
	getS3Store := aws.FlagVar(fl)
	getGithub := github.FlagVar(fl)
	getIndex := index.FlagVar(fl)
	getNewsletter := mailchimp.FlagVar(fl)
	getGoogle := google.FlagVar(fl)

	return func(l common.Logger) (svc Service, err error) {
		// Get PostgreSQL URL from Heroku if possible, else get it from flag
		if err = heroku.Configure(l, map[string]string{
			"postgres":    "DATABASE_URL",
			"google-json": "ALMANACK_GOOGLE_JSON",
		}); err != nil {
			return
		}
		if err = flagext.MustHave(fl, "postgres"); err != nil {
			return
		}

		client := *http.DefaultClient
		if *cache {
			httpcache.SetRounderTripper(&client, l)
		}

		if svc.ContentStore, err = getGithub(l); err != nil {
			l.Printf("could not connect to Github: %v", err)
			return
		}

		is, fs := getS3Store(l)

		return Service{
			Logger:           l,
			Client:           &client,
			Queries:          pg,
			ContentStore:     svc.ContentStore,
			SlackClient:      slack.New(*slackURL, l),
			ImageStore:       is,
			FileStore:        fs,
			Indexer:          getIndex(l),
			NewletterService: getNewsletter(&client),
			gsvc:             getGoogle(l),
		}, nil
	}
}
