package almanack

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/apex/gateway"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/carlmjohnson/flagext"
	"github.com/peterbourgon/ff"
	"github.com/spotlightpa/almanack/internal/netlifyid"
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
	mux.HandleFunc("/api/healthcheck", a.hello)
	mux.Handle("/api/user-info",
		a.netlifyIdentityMiddleware(http.HandlerFunc(a.userInfo)),
	)
	return mux
}

func (a *app) hello(w http.ResponseWriter, r *http.Request) {
	a.Println("start hello")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		a.Printf("could not dump request: %v", err)
		return
	}
	w.Write(b)
}

func (a *app) userInfo(w http.ResponseWriter, r *http.Request) {
	a.Println("start userInfo")
	ctx := r.Context()
	userinfo := ctx.Value(netlifyidContextKey)
	a.jsonResponse(http.StatusOK, w, userinfo)
}

type ErrorResponse struct {
	Code    int
	Message string
}

type netlifyidContextType int

const netlifyidContextKey = iota

func (a *app) netlifyIdentityMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.Println("start netlifyIdentityMiddleware")
		ctx := r.Context()
		lc, ok := lambdacontext.FromContext(ctx)
		if !ok {
			a.jsonResponse(http.StatusForbidden, w,
				ErrorResponse{http.StatusForbidden, "no context given: is this localhost?"})
			return
		}
		encoded := lc.ClientContext.Custom["netlify"]
		jwtBytes, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			a.Printf("bad netlify context: %q", encoded)
			a.jsonResponse(http.StatusInternalServerError, w,
				ErrorResponse{http.StatusInternalServerError, "bad response from Netlify"})
			return
		}
		var netID netlifyid.JWT
		if err = json.Unmarshal(jwtBytes, &netID); err != nil {
			a.Printf("could not unmarshal ID: %q", jwtBytes)
			a.jsonResponse(http.StatusInternalServerError, w,
				ErrorResponse{http.StatusInternalServerError, "bad JSON from Netlify"})
			return
		}

		ctx = context.WithValue(ctx, netlifyidContextKey, netID)
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	}
}

func (a *app) jsonResponse(statusCode int, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		a.Printf("jsonResponse problem: %v", err)
	}
}
