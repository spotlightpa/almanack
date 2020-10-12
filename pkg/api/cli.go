package api

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/carlmjohnson/gateway"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/peterbourgon/ff/v2"

	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/common"
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

	tls12 := fl.Bool("disable-tls-13", true, "set max TLS version to 1.2")
	fl.StringVar(&app.srcFeedURL, "src-feed", "", "source `URL` for Arc feed")
	fl.BoolVar(&app.isLambda, "lambda", false, "use AWS Lambda rather than HTTP")
	fl.StringVar(&app.port, "port", ":3001", "listen on port (HTTP only)")
	fl.StringVar(&app.mailchimpSignupURL, "mc-signup-url", "http://example.com", "`URL` to redirect users to for MailChimp signup")
	app.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	fl.Var(
		flagext.Logger(app.Logger, flagext.LogSilent),
		"silent",
		`don't log debug output`,
	)
	mcAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp")
	mcListID := fl.String("mc-list-id", "", "List `ID` MailChimp campaign")
	sentryDSN := fl.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), "almanack-api help\n\n")
		fl.PrintDefaults()
	}
	getService := almanack.Flags(fl)
	if err := ff.Parse(fl, args, ff.WithEnvVarPrefix("ALMANACK")); err != nil {
		return err
	}

	if err := app.initSentry(*sentryDSN, app.Logger); err != nil {
		return err
	}
	app.email = mailchimp.NewMailService(*mcAPIKey, *mcListID, app.Logger)
	app.auth = netlifyid.NewService(app.isLambda, app.Logger)
	var err error
	if app.svc, err = getService(app.Logger); err != nil {
		return err
	}
	if *tls12 {
		dt := http.DefaultTransport.(*http.Transport)
		dt.TLSClientConfig = &tls.Config{
			MaxVersion: tls.VersionTLS12,
		}
	}
	return nil
}

type appEnv struct {
	srcFeedURL         string
	port               string
	isLambda           bool
	mailchimpSignupURL string
	*log.Logger
	auth  common.AuthService
	email common.EmailService
	svc   almanack.Service
}

func (app *appEnv) exec() error {
	app.Printf("starting %s (%s)", AppName, almanack.BuildVersion)
	routes := sentryhttp.
		New(sentryhttp.Options{
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
			Repanic:         !app.isLambda,
		}).
		Handle(app.routes())

	if app.isLambda {
		var host string
		{
			u, _ := url.Parse(almanack.DeployURL)
			host = u.Hostname()
		}
		app.Printf("starting on AWS Lambda for %s", host)
		return gateway.ListenAndServe(host, routes)
	}

	app.Printf("starting on port %s", app.port)

	return http.ListenAndServe(app.port, routes)
}

func (app *appEnv) initSentry(dsn string, l common.Logger) error {
	var transport sentry.Transport
	if app.isLambda {
		l.Printf("setting sentry sync with timeout")
		transport = &sentry.HTTPSyncTransport{Timeout: 5 * time.Second}
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:       dsn,
		Release:   almanack.BuildVersion,
		Transport: transport,
	})
}
