package almanack

import (
	"errors"
	"flag"
	"net/http"

	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/herokuapi"
	"github.com/spotlightpa/almanack/internal/httpcache"
	"github.com/spotlightpa/almanack/internal/index"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/pkg/common"
)

func Flags(fl *flag.FlagSet) func(common.Logger) (svc Service, err error) {
	cache := fl.Bool("cache", false, "use in-memory cache for http requests")
	pg := db.FlagVar(fl, "postgres", "PostgreSQL database `URL`")
	slackURL := fl.String("slack-social-url", "", "Slack hook endpoint `URL` for social")
	checkHerokuPG := herokuapi.FlagVar(fl, "postgres")
	getS3Store := aws.FlagVar(fl)
	getGithub := github.FlagVar(fl)
	getIndex := index.FlagVar(fl)

	return func(l common.Logger) (svc Service, err error) {
		// Get PostgreSQL URL from Heroku if possible, else get it from flag
		if usedHeroku, err2 := checkHerokuPG(); err2 != nil {
			err = err2
			return
		} else if usedHeroku {
			l.Printf("got credentials from Heroku API")
		} else {
			l.Printf("did not get credentials from Heroku API")
		}

		if *pg == nil {
			err = errors.New("must set postgres URL")
			l.Printf("starting up: %v", err)
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
			Logger:       l,
			Client:       &client,
			Querier:      *pg,
			ContentStore: svc.ContentStore,
			SlackClient:  slack.New(*slackURL, l),
			ImageStore:   is,
			FileStore:    fs,
			Indexer:      getIndex(l),
		}, nil
	}
}
