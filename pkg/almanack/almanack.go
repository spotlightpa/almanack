package almanack

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

	"github.com/apex/gateway"
	"github.com/carlmjohnson/flagext"
	"github.com/peterbourgon/ff"
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

func parseArgs(args []string) (*app, error) {
	var a app
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fl.BoolVar(&a.useAWS, "lambda", false, "use AWS Lambda rather than HTTP")
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

	return &a, nil
}

type app struct {
	useAWS     bool
	port       string
	srcFeedURL string
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
	mux.Handle("/api/upcoming",
		a.netlifyPermissionMiddleware("editor", http.HandlerFunc(a.upcoming)),
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

const adminRole = "admin"

func (a *app) netlifyPermissionMiddleware(role string, next http.Handler) http.HandlerFunc {
	var inner http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		a.Println("starting permission middleware")
		if !a.useAWS {
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
		hasRole := false
		for _, r := range userinfo.User.AppMetadata.Roles {
			if r == role || r == adminRole {
				hasRole = true
				break
			}
		}
		a.Printf("permission middleware has role: %t", hasRole)
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
	}
	return a.netlifyIdentityMiddleware(inner)
}

func (a *app) fetchJSON(ctx context.Context, method, url string, v interface{}) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return errutil.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal error",
			Log:        fmt.Sprintf("bad downstream request: %v", err),
		}
	}
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
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

const (
	statusReady     = 5
	statusPublished = 6
)

func (a *app) upcoming(w http.ResponseWriter, r *http.Request) {
	a.Println("start upcoming")
	a.Printf("fetching %s", a.srcFeedURL)
	var feed jsonschema.API
	if err := a.fetchJSON(r.Context(), http.MethodGet, a.srcFeedURL, &feed); err != nil {
		a.errorResponse(w, err)
		return
	}
	// Filter out sub-drafts
	{
		i := 0
		for _, c := range feed.Contents {
			if c.Workflow.StatusCode >= statusReady {
				feed.Contents[i] = c
				i++
			}
		}
		feed.Contents = feed.Contents[:i]
	}
	a.jsonResponse(http.StatusOK, w, feed)
}
