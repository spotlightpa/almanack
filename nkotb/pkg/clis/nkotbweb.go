package clis

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/carlmjohnson/gateway"
	"github.com/carlmjohnson/versioninfo"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/spotlightpa/nkotb/build"
)

const NKOTBWebApp = "nkotbweb"

func NKOTBWeb(args []string) error {
	var app nkotbWebAppEnv
	err := app.ParseArgs(args)
	if err != nil {
		return err
	}
	err = app.Exec()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func (app *nkotbWebAppEnv) ParseArgs(args []string) error {
	fl := flag.NewFlagSet(NKOTBWebApp, flag.ContinueOnError)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), "%s - %s\n\n", NKOTBWebApp, versioninfo.Version)
		fl.PrintDefaults()
	}
	fl.IntVar(&app.port, "port", -1, "specify a port to use http rather than AWS Lambda")
	sentryDSN := fl.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	fl.StringVar(&app.oauthClientID, "client-id", "", "Google `Oauth client ID`")
	fl.StringVar(&app.oauthClientSecret, "client-secret", "", "Google `Oauth client secret`")
	secret := fl.String("signing-secret", "", "`secret` for HMAC cookie signing")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagext.ParseEnv(fl, NKOTBWebApp); err != nil {
		return err
	}
	if err := app.initSentry(*sentryDSN); err != nil {
		return err
	}
	app.signingSecret = []byte(*secret)
	logger.SetPrefix(NKOTBWebApp + " ")
	logger.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}

type nkotbWebAppEnv struct {
	port              int
	oauthClientID     string
	oauthClientSecret string
	signingSecret     []byte
}

func (app *nkotbWebAppEnv) Exec() (err error) {
	listener := gateway.ListenAndServe
	var portStr string
	if app.isLambda() {
		portStr = build.URL.Hostname()
	} else {
		portStr = fmt.Sprintf(":%d", app.port)
		build.URL.Host += portStr
		listener = http.ListenAndServe
	}
	routes := sentryhttp.
		New(sentryhttp.Options{
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
			Repanic:         !app.isLambda(),
		}).
		Handle(app.routes())

	logger.Printf("starting on %s", portStr)
	return listener(portStr, routes)
}

func (app *nkotbWebAppEnv) initSentry(dsn string) error {
	var transport sentry.Transport
	if app.isLambda() {
		logger.Printf("setting sentry sync with timeout")
		transport = &sentry.HTTPSyncTransport{Timeout: 5 * time.Second}
	}
	if dsn == "" {
		logger.Printf("no Sentry DSN")
		return nil
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:       dsn,
		Release:   build.Rev,
		Transport: transport,
	})
}

func (app *nkotbWebAppEnv) isLambda() bool {
	return app.port == -1
}
