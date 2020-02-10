package api

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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

	"github.com/spotlightpa/almanack/internal/arcjson"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/errutil"
	"github.com/spotlightpa/almanack/internal/filestore"
	"github.com/spotlightpa/almanack/internal/herokuapi"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/redis"
	"github.com/spotlightpa/almanack/internal/redisflag"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

const AppName = "almanack-api"

func CLI(args []string) error {
	a, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Startup error: %v\n", err)
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
	fl.StringVar(&a.port, "port", ":3001", "listen on port (HTTP only)")
	getDialer := redisflag.Var(fl, "redis-url", "`URL` connection string for Redis")
	a.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	fl.Var(
		flagext.Logger(a.Logger, flagext.LogSilent),
		"silent",
		`don't log debug output`,
	)
	checkHeroku := herokuapi.FlagVar(fl)
	getImageStore := aws.FlagVar(fl)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), "almanack-api help\n\n")
		fl.PrintDefaults()
	}
	if err := ff.Parse(fl, args, ff.WithEnvVarPrefix("ALMANACK")); err != nil {
		return nil, err
	}
	// Get Redis URL from Heroku if possible, else get it from config, else use files
	if connURL, err := checkHeroku(); err != nil {
		return nil, err
	} else if connURL != "" {
		a.Logger.Printf("got credentials from Heroku")
		dialer, err := redisflag.Parse(connURL)
		if err != nil {
			return nil, err
		}
		if a.store, err = redis.New(dialer, a.Logger); err != nil {
			return nil, err
		}
	} else {
		if d := getDialer(); d != nil {
			a.Logger.Printf("got Redis URL directly")
			var err error
			if a.store, err = redis.New(d, a.Logger); err != nil {
				return nil, err
			}
		} else {
			a.Logger.Printf("using filestore")
			a.store = filestore.New("", "almanack", a.Logger)
		}
	}
	a.imageStore = getImageStore(a.Logger)
	a.auth = netlifyid.NewService(a.isLambda, a.Logger)
	a.c = http.DefaultClient

	return &a, nil
}

type appEnv struct {
	port       string
	isLambda   bool
	c          *http.Client
	auth       almanack.AuthService
	store      almanack.DataStore
	imageStore almanack.ImageStore
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
	r.Get("/api/healthcheck", a.ping)
	r.Route("/api", func(r chi.Router) {
		r.Use(a.authMiddleware)
		r.Get("/user-info", a.userInfo)
		r.With(
			a.hasRoleMiddleware("editor"),
		).Get("/upcoming", a.upcoming)
		r.With(
			a.hasRoleMiddleware("Spotlight PA"),
		).Group(func(r chi.Router) {
			r.Get("/articles/{id}", a.getArticle)
			r.Post("/articles/{id}", a.postArticle)
			r.Post("/get-signed-upload", a.getSignedUpload)
		})
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
	if !errors.As(err, &errResp) {
		errResp.StatusCode = http.StatusInternalServerError
		errResp.Message = "internal error"
		errResp.Log = err.Error()
	}
	a.Println(errResp.Log)
	a.jsonResponse(errResp.StatusCode, w, errResp)
}

func (a *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	a.Println("start ping")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		a.errorResponse(w, err)
		return
	}
	w.Write(b)
}

func (a *appEnv) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Println("start authMiddleware")
		r, err := a.auth.AddToRequest(r)
		if err != nil {
			a.errorResponse(w, err)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (a *appEnv) userInfo(w http.ResponseWriter, r *http.Request) {
	a.Println("start userInfo")
	userinfo, err := netlifyid.FromRequest(r)
	if err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusOK, w, userinfo)
}

func (a *appEnv) hasRoleMiddleware(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.Println("starting hasRoleMiddleware")
			if err := a.auth.HasRole(r, role); err != nil {
				a.errorResponse(w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

const feedKey = "almanack-worker.feed"

func (a *appEnv) upcoming(w http.ResponseWriter, r *http.Request) {
	a.Println("start upcoming")

	var feed arcjson.API
	if err := a.store.Get(feedKey, &feed); err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusOK, w, feed)
}

func (a *appEnv) getArticle(w http.ResponseWriter, r *http.Request) {
	a.Println("start getArticle")

	articleID := chi.URLParam(r, "id")

	arcsvc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	sas := almanack.ScheduledArticleService{
		ArticleService: arcsvc,
		DataStore:      a.store,
		Logger:         a.Logger,
	}

	article, err := sas.Get(articleID)
	if err != nil {
		a.errorResponse(w, err)
		return
	}

	a.jsonResponse(http.StatusOK, w, article)
}

func (a *appEnv) postArticle(w http.ResponseWriter, r *http.Request) {
	a.Println("start postArticle")

	articleID := chi.URLParam(r, "id")

	var userData almanack.ScheduledArticle
	if err := errutil.DecodeJSONBody(w, r, &userData); err != nil {
		a.errorResponse(w, err)
		return
	}

	// Get the lock
	unlock, err := a.store.GetLock("almanack.scheduled-articles-lock")
	defer unlock()
	if err != nil {
		a.errorResponse(w, err)
		return
	}

	// Save the article
	if err := a.store.Set("almanack.scheduled-article."+articleID, &userData); err != nil {
		a.errorResponse(w, err)
		return
	}

	// Get the existing list of scheduled articles
	ids := map[string]bool{}
	if err = a.store.Get("almanack.scheduled-articles-list", &ids); err != nil &&
		!errors.Is(err, errutil.NotFound) {
		a.errorResponse(w, err)
		return
	}

	// If the status of the article changed, update the list
	shouldPub := userData.ScheduleFor != nil
	hasChanged := shouldPub != ids[articleID]

	if hasChanged {
		ids[articleID] = shouldPub
		if err := a.store.Set("almanack.scheduled-articles-list", &ids); err != nil {
			a.errorResponse(w, err)
			return
		}
	}

	a.jsonResponse(http.StatusAccepted, w, &struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		http.StatusAccepted,
		http.StatusText(http.StatusAccepted),
	})
}

func (a *appEnv) getSignedUpload(w http.ResponseWriter, r *http.Request) {
	signedURL, filename, err := a.imageStore.GetSignedUpload()
	if err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusOK, w, &struct {
		SignedURL string `json:"signed-url"`
		FileName  string `json:"filename"`
	}{
		signedURL,
		filename,
	})
}
