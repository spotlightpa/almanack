package almanack

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/apex/gateway"
	"github.com/auth0-community/go-auth0"
	_ "github.com/auth0/go-jwt-middleware"
	"github.com/carlmjohnson/flagext"
	"github.com/peterbourgon/ff"
	"gopkg.in/square/go-jose.v2"
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
	fl.BoolVar(&a.useAWS, "lambda", false, "use AWS Lambda rather than HTTP")
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

	return &a, nil
}

type app struct {
	useAWS bool
	port   string
	*log.Logger
}

func (a *app) exec() error {
	listener := http.ListenAndServe
	if a.useAWS {
		a.Printf("starting on AWS Lambda")
		listener = gateway.ListenAndServe
	} else {
		a.Printf("starting on port %s", a.port)
	}
	return listener(a.port, a.routes())
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/healthcheck", a.xxx())
	return mux
}

func (a *app) hello(w http.ResponseWriter, r *http.Request) {
	a.Println("hello start")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		a.Printf("could not dump request: %v", err)
		return
	}
	w.Write(b)
}

func (a *app) xxx() http.HandlerFunc {
	// jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
	// 	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
	// 		return []byte("My Secret"), nil
	// 	},
	// 	SigningMethod: jwt.SigningMethodHS256,
	// })

	// app := jwtMiddleware.Handler(myHandler)

	secretProvider := auth0.NewJWKClient(auth0.JWKClientOptions{
		URI: "https://dev-o74bq264.auth0.com/.well-known/jwks.json",
	}, nil)
	configuration := auth0.NewConfiguration(secretProvider, []string{"https://spotlighpa-almanack.netlify.com"}, "", jose.RS256)
	validator := auth0.NewValidator(configuration, nil)
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := validator.ValidateRequest(r)
		a.Printf("token: %v err: %v", t, err)
		a.writeJSON(http.StatusOK, w, t)
	}
}

func (a *app) writeJSON(statusCode int, w http.ResponseWriter, data interface{}) {
	w.WriteHeader(statusCode)
	e := json.NewEncoder(w)
	if err := e.Encode(data); err != nil {
		a.Printf("could not send JSON: %v", err)
	}
}
