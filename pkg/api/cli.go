package api

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/peterbourgon/ff/v2"
	"github.com/piotrkubisa/apigo"

	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/github"
	"github.com/spotlightpa/almanack/internal/herokuapi"
	"github.com/spotlightpa/almanack/internal/httpcache"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

const AppName = "almanack-api"

func CLI(args []string) error {
	var app appEnv
	if err := app.parseArgs(args); err != nil {
		fmt.Fprintf(os.Stderr, "Startup error: %v\n", err)
		return err
	}
	if err := app.exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return err
	}
	return nil
}

func (app *appEnv) parseArgs(args []string) error {
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)

	pg := db.FlagVar(fl, "postgres", "PostgreSQL database `URL`")
	fl.StringVar(&app.srcFeedURL, "src-feed", "", "source `URL` for Arc feed")
	cache := fl.Bool("cache", false, "use in-memory cache for fetched JSON")
	fl.BoolVar(&app.isLambda, "lambda", false, "use AWS Lambda rather than HTTP")
	fl.StringVar(&app.port, "port", ":3001", "listen on port (HTTP only)")
	fl.StringVar(&app.mailchimpSignupURL, "mc-signup-url", "http://example.com", "`URL` to redirect users to for MailChimp signup")
	checkHerokuPG := herokuapi.FlagVar(fl, "postgres")
	app.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	fl.Var(
		flagext.Logger(app.Logger, flagext.LogSilent),
		"silent",
		`don't log debug output`,
	)
	getImageStore := aws.FlagVar(fl)
	mcAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp")
	mcListID := fl.String("mc-list-id", "", "List `ID` MailChimp campaign")
	sentryDSN := fl.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	getGithub := github.FlagVar(fl)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), "almanack-api help\n\n")
		fl.PrintDefaults()
	}
	if err := ff.Parse(fl, args, ff.WithEnvVarPrefix("ALMANACK")); err != nil {
		return err
	}

	if err := app.initSentry(*sentryDSN); err != nil {
		return err
	}

	// Get PostgreSQL URL from Heroku if possible, else get it from flag
	if usedHeroku, err := checkHerokuPG(); err != nil {
		return err
	} else if usedHeroku {
		app.Logger.Printf("got credentials from Heroku")
	} else {
		app.Logger.Printf("did not get credentials Heroku")
	}

	if *pg == nil {
		err := errors.New("must set postgres URL")
		app.Logger.Printf("starting up: %v", err)
		return err
	}

	app.email = mailchimp.NewMailService(*mcAPIKey, *mcListID, app.Logger)
	app.imageStore = getImageStore(app.Logger)
	app.auth = netlifyid.NewService(app.isLambda, app.Logger)
	app.c = http.DefaultClient
	if *cache {
		httpcache.SetRounderTripper(app.c, app.Logger)
	}
	if gh, err := getGithub(app.Logger); err != nil {
		app.Logger.Printf("could not connect to Github: %v", err)
		return err
	} else {
		app.gh = gh
	}
	app.svc = almanack.Service{
		Querier:      *pg,
		Logger:       app.Logger,
		ContentStore: app.gh,
		ImageStore:   app.imageStore,
		Client:       app.c,
	}

	return nil
}

type appEnv struct {
	srcFeedURL         string
	port               string
	isLambda           bool
	mailchimpSignupURL string
	c                  *http.Client
	auth               almanack.AuthService
	gh                 almanack.ContentStore
	imageStore         almanack.ImageStore
	email              almanack.EmailService
	svc                almanack.Service
	*log.Logger
}

func (app *appEnv) exec() error {
	app.Printf("starting %s (%s)", AppName, almanack.BuildVersion)

	listener := http.ListenAndServe
	if app.isLambda {
		app.Printf("starting on AWS Lambda")
		apigo.ListenAndServe("", app.routes())
		panic("unreachable")
	}

	app.Printf("starting on port %s", app.port)
	routes := sentryhttp.
		New(sentryhttp.Options{
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
		}).
		Handle(app.routes())

	return listener(app.port, routes)
}

func (app *appEnv) initSentry(dsn string, l almanack.Logger) error {
	var transport sentry.Transport
	if app.isLambda {
		l.Printf("setting sentry timeout %q", dsn)
		transport = &sentry.HTTPSyncTransport{Timeout: 5 * time.Second}
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:       dsn,
		Release:   almanack.BuildVersion,
		Transport: transport,
	})
}
