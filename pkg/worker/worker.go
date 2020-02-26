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
	var app appEnv
	err := app.parseArgs(args)
	if err != nil {
		app.sc.Post(
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
	if err := app.exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		app.sc.Post(
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

func (app *appEnv) parseArgs(args []string) error {
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fl.StringVar(&app.srcFeedURL, "src-feed", "", "source `URL` for Arc feed")
	mcAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp")
	mcListID := fl.String("mc-list-id", "", "List `ID` MailChimp campaign")
	getDialer := redisflag.Var(fl, "redis-url", "`URL` connection string for Redis")
	slackURL := fl.String("slack-hook-url", "", "Slack hook endpoint `URL`")
	app.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	fl.Var(
		flagext.Logger(app.Logger, flagext.LogSilent),
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
		return err
	}
	app.sc = slack.New(*slackURL, app.Logger)

	app.email = mailchimp.NewMailService(*mcAPIKey, *mcListID, app.Logger)
	if d := getDialer(); d != nil {
		var err error
		if app.store, err = redis.New(d, app.Logger); err != nil {
			app.Logger.Printf("could not connect to redis: %v", err)
			return err
		}
	} else {
		app.store = filestore.New("", "almanack", app.Logger)
	}

	if gh, err := getGithub(app.Logger); err != nil {
		app.Logger.Printf("could not connect to Github: %v", err)
		return err
	} else {
		app.gh = gh
	}

	return nil
}

type appEnv struct {
	srcFeedURL string
	store      almanack.DataStore
	email      almanack.EmailService
	gh         almanack.ContentStore
	sc         slack.Client
	*log.Logger
}

func (app *appEnv) exec() error {
	app.Println("starting", AppName)
	start := time.Now()
	defer func() { app.Println("finished in", time.Since(start)) }()

	return errutil.ExecParallel(
		app.updateFeed,
		app.publishStories,
	)
}

func (app *appEnv) updateFeed() error {
	app.Println("starting updateFeed")
	if app.srcFeedURL == "" {
		app.Println("aborting: no feed URL provided")
		return nil
	}
	app.Println("fetching", app.srcFeedURL)
	var newfeed arcjson.API
	if err := httpjson.Get(context.Background(), nil, app.srcFeedURL, &newfeed); err != nil {
		return err
	}

	svc := arcjson.FeedService{DataStore: app.store, Logger: app.Logger}
	if err := svc.StoreFeed(newfeed); err != nil {
		return err
	}

	return nil
}

func (app *appEnv) publishStories() error {
	app.Println("starting publishStories")
	sas := almanack.ScheduledArticleService{
		DataStore: app.store,
		Logger:    app.Logger,
	}

	return sas.PopScheduledArticles(func(articles []*almanack.ScheduledArticle) error {
		for _, article := range articles {
			ctx := context.Background()
			if err := article.Publish(ctx, app.gh); err != nil {
				return err
			}
		}
		return nil
	})
}
