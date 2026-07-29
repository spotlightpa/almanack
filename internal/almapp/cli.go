package almapp

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/carlmjohnson/gateway"
	"github.com/earthboundkid/flagx/v2"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/getsentry/sentry-go"

	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
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
	getService := almsvc.AddFlags(fl)

	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagx.ParseEnv(fl, "almanack"); err != nil {
		return err
	}
	if app.svc.IsLambda {
		almlog.UseLambdaLogger()
	} else {
		almlog.UseDevLogger()
	}
	if err := app.initSentry(*sentryDSN); err != nil {
		return err
	}
	var err error
	if app.svc, err = getService(); err != nil {
		return err
	}
	return nil
}

type appEnv struct {
	port string
	svc  almsvc.Services
}

func (app *appEnv) exec() error {
	routes := app.routes()

	var host string
	if app.svc.IsLambda {
		u, _ := url.Parse(almsvc.DeployURL)
		host = u.Hostname()
	}
	almlog.Logger.Info("appEnv.exec",
		"app", AppName,
		"version", versioninfo.Short(),
		"is-lambda", app.svc.IsLambda,
		"host", host,
		"port", app.port,
	)
	if app.svc.IsLambda {
		return gateway.ListenAndServe(host, routes)
	}

	return http.ListenAndServe(app.port, routes)
}

func (app *appEnv) initSentry(dsn string) error {
	var transport sentry.Transport
	if app.svc.IsLambda {
		almlog.Logger.Debug("initSentry", "sync", true, "timeout", 5*time.Second)
		transport = &sentry.HTTPSyncTransport{Timeout: 5 * time.Second}
	} else {
		almlog.Logger.Debug("initSentry", "sync", false, "timeout", false)
	}
	return sentry.Init(sentry.ClientOptions{
		Dsn:       dsn,
		Release:   almsvc.BuildVersion,
		Transport: transport,
	})
}
