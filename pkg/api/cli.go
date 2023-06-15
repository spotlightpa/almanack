package api

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/carlmjohnson/flagx"
	"github.com/carlmjohnson/gateway"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"golang.org/x/exp/slog"

	"github.com/spotlightpa/almanack/pkg/almlog"
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
	fl.Func("level", "log level", func(s string) error {
		l, _ := strconv.Atoi(s)
		almlog.Level.Set(slog.Level(l))
		return nil
	})
	sentryDSN := fl.String("sentry-dsn", "", "DSN `pseudo-URL` for Sentry")
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), "almanack-api help\n\n")
		fl.PrintDefaults()
	}
	// getService := almanack.AddFlags(fl)

	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagx.ParseEnv(fl, "almanack"); err != nil {
		return err
	}
	if app.isLambda {
		almlog.UseLambdaLogger()
	} else {
		almlog.UseDevLogger()
	}
	if err := app.initSentry(*sentryDSN); err != nil {
		return err
	}
	// app.auth = netlifyid.NewService(app.isLambda)
	// var err error
	// if app.svc, err = getService(); err != nil {
	// 	return err
	// }
	return nil
}

type appEnv struct {
	port     string
	isLambda bool
	// auth     netlifyid.AuthService
	// svc      almanack.Services
}

func (app *appEnv) exec() error {
	routes := sentryhttp.
		New(sentryhttp.Options{
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
			Repanic:         !app.isLambda,
		}).
		Handle(app.routes())

	var host string
	// if app.isLambda {
	// 	u, _ := url.Parse(almanack.DeployURL)
	// 	host = u.Hostname()
	// }
	// almlog.Logger.Info("appEnv.exec",
	// 	"app", AppName,
	// 	"version", versioninfo.Short(),
	// 	"is-lambda", app.isLambda,
	// 	"host", host,
	// 	"port", app.port,
	// )
	if app.isLambda {
		return gateway.ListenAndServe(host, routes)
	}

	return http.ListenAndServe(app.port, routes)
}

func (app *appEnv) initSentry(dsn string) error {
	var transport sentry.Transport
	if app.isLambda {
		almlog.Logger.Debug("initSentry", "sync", true, "timeout", 5*time.Second)
		transport = &sentry.HTTPSyncTransport{Timeout: 5 * time.Second}
	} else {
		almlog.Logger.Debug("initSentry", "sync", false, "timeout", false)
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
		// Release:   almanack.BuildVersion,
		Transport: transport,
	})
}
