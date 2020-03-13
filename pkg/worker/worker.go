package worker

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/getsentry/sentry-go"
	"github.com/peterbourgon/ff/v2"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/httpjson"
	"github.com/spotlightpa/almanack/internal/mailchimp"
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
		sentry.CaptureException(err)
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
	pg := db.FlagVar(fl, "postgres", "PostgreSQL database `URL`")
	mcAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp")
	mcListID := fl.String("mc-list-id", "", "List `ID` MailChimp campaign")
	slackURL := fl.String("slack-hook-url", "", "Slack hook endpoint `URL`")
	sentryDSN := fl.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
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
	if err := flagext.MustHave(fl, "postgres"); err != nil {
		return err
	}

	app.sc = slack.New(*slackURL, app.Logger)

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:     *sentryDSN,
		Release: almanack.BuildVersion,
	}); err != nil {
		return err
	}

	app.db = *pg

	app.email = mailchimp.NewMailService(*mcAPIKey, *mcListID, app.Logger)

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
	email      almanack.EmailService
	gh         almanack.ContentStore
	db         db.Querier
	sc         slack.Client
	*log.Logger
}

func (app *appEnv) exec() error {
	app.Printf("starting %s (%s)", AppName, almanack.BuildVersion)
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
	var newfeed almanack.ArcAPI
	if err := httpjson.Get(context.Background(), nil, app.srcFeedURL, &newfeed); err != nil {
		return err
	}

	svc := almanack.FeedService{
		Querier: app.db,
		Logger:  app.Logger,
	}
	if err := svc.StoreFeed(context.Background(), newfeed, false); err != nil {
		return err
	}

	return nil
}

func (app *appEnv) publishStories() error {
	app.Println("starting publishStories")
	sas := almanack.ScheduledArticleService{
		Querier: app.db,
		Logger:  app.Logger,
	}

	ctx := context.Background()
	return sas.PopScheduledArticles(ctx, func(articles []*almanack.ScheduledArticle) error {
		for _, article := range articles {
			if err := article.Publish(ctx, app.gh); err != nil {
				return err
			}
		}
		return nil
	})
}
