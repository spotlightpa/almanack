package api

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/peterbourgon/ff"
	"github.com/piotrkubisa/apigo"
	"golang.org/x/xerrors"

	"github.com/spotlightpa/almanack/internal/errutil"
	"github.com/spotlightpa/almanack/internal/jsonschema"
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

func parseArgs(args []string) (*appEnv, error) {
	var a appEnv
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fl.BoolVar(&a.isLambda, "lambda", false, "use AWS Lambda rather than HTTP")
	cache := fl.Bool("cache", false, "use in-memory cache for fetched JSON")
	fl.StringVar(&a.port, "port", ":3001", "listen on port (HTTP only)")
	fl.StringVar(&a.srcFeedURL, "src-feed", "", "source URL for Arc feed")
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

	a.c = http.DefaultClient
	if *cache {
		SetRounderTripper(a.c, a.Logger)
	}

	return &a, nil
}

type appEnv struct {
	isLambda   bool
	port       string
	srcFeedURL string
	c          *http.Client
	*log.Logger
}

func (a *appEnv) exec() error {
	listener := http.ListenAndServe
	if a.isLambda {
		a.Printf("starting on AWS Lambda")
		apigo.ListenAndServe("", a.routes())
		panic("unreachable")
	}

	a.Printf("starting on port %s", a.port)
	return listener(a.port, a.routes())
}

func (a *appEnv) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: a.Logger}))
	r.Use(middleware.Recoverer)
	r.Get("/api/healthcheck", a.hello)
	r.Route("/api", func(r chi.Router) {
		r.Use(a.netlifyIdentityMiddleware)
		r.Get("/user-info", a.userInfo)
		r.With(
			a.netlifyPermissionMiddleware("editor"),
		).Get("/upcoming", a.upcoming)
	})
	return r
}

func (a *appEnv) loggingMiddleware(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		a.Printf("request took %v", time.Since(start))
	}
}

func (a *appEnv) jsonResponse(statusCode int, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		a.Printf("jsonResponse problem: %v", err)
	}
}

func (a *appEnv) errorResponse(w http.ResponseWriter, err error) {
	var errResp errutil.Response
	if !xerrors.As(err, &errResp) {
		errResp.StatusCode = http.StatusInternalServerError
		errResp.Message = "internal error"
		errResp.Log = err.Error()
	}
	a.Println(errResp.Log)
	a.jsonResponse(errResp.StatusCode, w, errResp)
}

func (a *appEnv) hello(w http.ResponseWriter, r *http.Request) {
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

func (a *appEnv) netlifyIdentityMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Println("start netlifyIdentityMiddleware")
		if !a.isLambda {
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
	})
}

func (a *appEnv) userInfo(w http.ResponseWriter, r *http.Request) {
	a.Println("start userInfo")
	userinfo := getNetlifyID(r)
	a.jsonResponse(http.StatusOK, w, userinfo)
}

func (a *appEnv) netlifyPermissionMiddleware(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.Println("starting permission middleware")
			if !a.isLambda {
				a.Println("skipping permission middleware")
				next.ServeHTTP(w, r)
				return
			}

			userinfo := getNetlifyID(r)
			if userinfo == nil {
				err := errutil.Response{
					StatusCode: http.StatusInternalServerError,
					Message:    "user info not set",
					Log:        "no user info: is this localhost?",
				}
				a.errorResponse(w, err)
				return
			}
			hasRole := userinfo.HasRole(role)
			a.Printf("permission middleware: %s has role %s == %t",
				userinfo.User.Email, role, hasRole)
			if !hasRole {
				err := errutil.Response{
					StatusCode: http.StatusForbidden,
					Message:    http.StatusText(http.StatusForbidden),
					Log: fmt.Sprintf(
						"unauthorized user only had roles: %v",
						userinfo.User.AppMetadata.Roles),
				}
				a.errorResponse(w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (a *appEnv) fetchJSON(ctx context.Context, method, url string, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return errutil.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal error",
			Log:        fmt.Sprintf("bad downstream request: %v", err),
		}
	}
	resp, err := a.c.Do(req)
	if err != nil {
		return errutil.Response{
			StatusCode: http.StatusBadGateway,
			Message:    "could not contact Inquirer server",
			Log:        fmt.Sprintf("bad downstream connect: %v", err),
		}
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errutil.Response{
			StatusCode: http.StatusBadGateway,
			Message:    "could not read from Inquirer server",
			Log:        fmt.Sprintf("bad downstream read: %v", err),
		}
	}

	if err = json.Unmarshal(b, v); err != nil {
		return errutil.Response{
			StatusCode: http.StatusBadGateway,
			Message:    "could not decode from Inquirer server",
			Log:        fmt.Sprintf("bad downstream decode: %v", err),
		}
	}

	return nil
}

func (a *appEnv) upcoming(w http.ResponseWriter, r *http.Request) {
	a.Println("start upcoming")
	a.Printf("fetching %s", a.srcFeedURL)
	var feed jsonschema.API
	if err := a.fetchJSON(r.Context(), http.MethodGet, a.srcFeedURL, &feed); err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusOK, w, feed)
}
