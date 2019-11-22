package almanack

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/apex/gateway"
	"github.com/carlmjohnson/flagext"
	"github.com/peterbourgon/ff"
	"golang.org/x/xerrors"

	"github.com/spotlightpa/almanack/internal/errutil"
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

func (a *app) jsonResponse(statusCode int, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		a.Printf("jsonResponse problem: %v", err)
	}
}

func (a *app) errorResponse(w http.ResponseWriter, err error) {
	var errResp errutil.Response
	if !xerrors.As(err, &errResp) {
		errResp.StatusCode = http.StatusInternalServerError
		errResp.Message = "internal error"
		errResp.Log = err.Error()
	}
	a.Println(errResp.Log)
	a.jsonResponse(errResp.StatusCode, w, errResp)
}

func (a *app) hello(w http.ResponseWriter, r *http.Request) {
	a.Println("start hello")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		a.errorResponse(w, err)
		return
	}
	w.Write(b)
}

type netlifyidContextType int

const netlifyidContextKey = iota

func setNetlifyID(r *http.Request, netID *netlifyid.JWT) *http.Request {
	ctx := context.WithValue(r.Context(), netlifyidContextKey, netID)
	return r.WithContext(ctx)
}

func getNetlifyID(r *http.Request) *netlifyid.JWT {
	ctx := r.Context()
	val := ctx.Value(netlifyidContextKey)
	if val == nil { // interface nil
		return nil // *JWT nil
	}
	return val.(*netlifyid.JWT)
}

func (a *app) netlifyIdentityMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		a.Println("start netlifyIdentityMiddleware")
		if !a.useAWS {
			a.Println("skip netlifyIdentityMiddleware")
			h.ServeHTTP(w, r)
			return
		}
		netID, err := netlifyid.FromRequest(r)
		if err != nil {
			a.errorResponse(w, err)
			return
		}
		r = setNetlifyID(r, netID)
		h.ServeHTTP(w, r)
	}
}

func (a *app) userInfo(w http.ResponseWriter, r *http.Request) {
	a.Println("start userInfo")
	userinfo := getNetlifyID(r)
	a.jsonResponse(http.StatusOK, w, userinfo)
}
