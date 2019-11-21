package almanack

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/apex/gateway"
	"github.com/carlmjohnson/flagext"
	"github.com/peterbourgon/ff"
)

const AppName = "almanack-api"

func CLI(args []string) error {
	a, err := parseArgs(args)
	if err != nil {
		return err
	}
	if err := a.exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return err
	}
	return nil
}

func parseArgs(args []string) (*app, error) {
	var a app
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	useAWS := flag.Bool("lambda", false, "use AWS Lambda rather than HTTP")
	fl.StringVar(&a.port, "port", ":3001", "listen on port (HTTP only)")
	a.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	fl.Var(
		flagext.Logger(a.Logger, flagext.LogSilent),
		"silent",
		`don't log debug output`,
	)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `almanack-api help`)
		fl.PrintDefaults()
	}
	if err := ff.Parse(fl, args, ff.WithEnvVarPrefix("ALMANACK")); err != nil {
		return nil, err
	}

	a.listener = http.ListenAndServe
	if *useAWS {
		a.listener = gateway.ListenAndServe
	}

	return &a, nil
}

type app struct {
	listener func(string, http.Handler) error
	port     string
	*log.Logger
}

func (a *app) exec() error {
	a.Printf("starting on port %s", a.port)
	return a.listener(a.port, a.routes())
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthcheck", a.hello)
	return mux
}

func (a *app) hello(w http.ResponseWriter, r *http.Request) {
	a.Printf("hello: %v", r)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		a.Printf("could not dump request: %v", err)
		return
	}
	w.Write(b)
}
