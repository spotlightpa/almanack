package worker

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/peterbourgon/ff"

	"github.com/spotlightpa/almanack/internal/arcjson"
	"github.com/spotlightpa/almanack/internal/filestore"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/httpjson"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/redis"
	"github.com/spotlightpa/almanack/internal/redisflag"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

const AppName = "almanack-worker"

func CLI(args []string) error {
	a, err := parseArgs(args)
	if err != nil {
		return err
	}
	if err := a.exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		a.sc.Post(
			slack.Message{
				Attachments: []slack.Attachment{
					{
						Title: "Almanack Worker Error",
						Text:  err.Error(),
						Color: "#da291c",
					}}},
		)

		return err
	}
	return nil
}

func parseArgs(args []string) (*appEnv, error) {
	var a appEnv
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fl.StringVar(&a.srcFeedURL, "src-feed", "", "source `URL` for Arc feed")
	mcAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp")
	mcListID := fl.String("mc-list-id", "", "List `ID` MailChimp campaign")
	getDialer := redisflag.Var(fl, "redis-url", "`URL` connection string for Redis")
	slackURL := fl.String("slack-hook-url", "", "Slack hook endpoint `URL`")
	a.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	fl.Var(
		flagext.Logger(a.Logger, flagext.LogSilent),
		"silent",
		`don't log debug output`,
	)
	getGithub := github.FlagVar(fl)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `almanack-worker help

Options:
`)
		fl.PrintDefaults()
	}
	if err := ff.Parse(fl, args, ff.WithEnvVarPrefix("ALMANACK")); err != nil {
		return nil, err
	}
	a.email = mailchimp.NewMailService(*mcAPIKey, *mcListID, a.Logger)
	if d := getDialer(); d != nil {
		var err error
		if a.store, err = redis.New(d, a.Logger); err != nil {
			a.Logger.Printf("could not connect to redis: %v", err)
			return nil, err
		}
	} else {
		a.store = filestore.New("", "almanack", a.Logger)
	}
	a.sc = slack.New(*slackURL, a.Logger)
	if gh, err := getGithub(a.Logger); err != nil {
		a.Logger.Printf("could not connect to Github: %v", err)
		return nil, err
	} else {
		a.gh = gh
	}
	return &a, nil
}

type appEnv struct {
	srcFeedURL string
	store      almanack.DataStore
	email      almanack.EmailService
	gh         almanack.ContentStore
	sc         slack.Client
	*log.Logger
}

func (a *appEnv) exec() error {
	a.Println("starting", AppName)
	start := time.Now()
	defer func() { a.Println("finished in", time.Since(start)) }()

	return errutil.ExecParallel(
		a.updateFeed,
		a.publishStories,
	)
}

func (a *appEnv) updateFeed() error {
	a.Println("starting updateFeed")
	if a.srcFeedURL == "" {
		a.Println("aborting: no feed URL provided")
		return nil
	}
	a.Println("fetching", a.srcFeedURL)
	var newfeed arcjson.API
	if err := httpjson.Get(context.Background(), nil, a.srcFeedURL, &newfeed); err != nil {
		return err
	}

	svc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	if err := svc.StoreFeed(newfeed); err != nil {
		return err
	}

	return nil
}

func (a *appEnv) publishStories() error {
	a.Println("starting publishStories")
	sas := almanack.ScheduledArticleService{
		DataStore: a.store,
		Logger:    a.Logger,
	}

	return sas.PopScheduledArticles(func(articles []*almanack.ScheduledArticle) error {
		for _, article := range articles {
			ctx := context.Background()
			msg := fmt.Sprintf("Content: publishing %q", article.ID)
			path := article.ContentFilepath()
			data, err := article.ToTOML()
			if err != nil {
				return err
			}
			if err = a.gh.CreateFile(ctx, msg, path, []byte(data)); err != nil {
				return err
			}
		}
		return nil
	})
}
