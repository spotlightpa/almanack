package api

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/carlmjohnson/flagx"
	"github.com/carlmjohnson/gateway"
	"github.com/carlmjohnson/versioninfo"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"

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

	fl.BoolVar(&app.isLambda, "lambda", false, "use AWS Lambda rather than HTTP")
	fl.StringVar(&app.port, "port", ":33160", "listen on port (HTTP only)")
	app.Logger = log.Default()
	app.Logger.SetPrefix(AppName + " ")
	flagx.BoolFunc(fl, "silent", "don't log debug output", func() error {
		app.Logger.SetOutput(io.Discard)
		return nil
	})
	sentryDSN := fl.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), "almanack-api help\n\n")
		fl.PrintDefaults()
	}
	getService := almanack.AddFlags(fl)

	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagx.ParseEnv(fl, "almanack"); err != nil {
		return err
	}
	if err := app.initSentry(*sentryDSN, app.Logger); err != nil {
		return err
	}
	app.auth = netlifyid.NewService(app.isLambda, app.Logger)
	var err error
	if app.svc, err = getService(app.Logger); err != nil {
		return err
	}
	return nil
}

type appEnv struct {
	port     string
	isLambda bool
	*log.Logger
	auth netlifyid.AuthService
	svc  almanack.Service
}

func (app *appEnv) exec() error {
	app.Printf("starting %s (%s)", AppName, versioninfo.Short())
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
