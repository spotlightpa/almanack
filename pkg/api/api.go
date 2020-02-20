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
	"github.com/spotlightpa/almanack/internal/filestore"
	"github.com/spotlightpa/almanack/internal/herokuapi"
	"github.com/spotlightpa/almanack/internal/httpcache"
	"github.com/spotlightpa/almanack/internal/httpjson"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/redis"
	"github.com/spotlightpa/almanack/internal/redisflag"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/errutil"
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
	fl.StringVar(&a.srcFeedURL, "src-feed", "", "source `URL` for Arc feed")
	cache := fl.Bool("cache", false, "use in-memory cache for fetched JSON")
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
	mcAPIKey := fl.String("mc-api-key", "", "API `key` for MailChimp")
	mcListID := fl.String("mc-list-id", "", "List `ID` MailChimp campaign")
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
	a.email = mailchimp.NewMailService(*mcAPIKey, *mcListID, a.Logger)
	a.imageStore = getImageStore(a.Logger)
	a.auth = netlifyid.NewService(a.isLambda, a.Logger)
	a.c = http.DefaultClient
	if *cache {
		httpcache.SetRounderTripper(a.c, a.Logger)
	}
	return &a, nil
}

type appEnv struct {
	srcFeedURL string
	port       string
	isLambda   bool
	c          *http.Client
	auth       almanack.AuthService
	store      almanack.DataStore
	imageStore almanack.ImageStore
	email      almanack.EmailService
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
		).Group(func(r chi.Router) {
			r.Get("/available-articles", a.listAvailable)
			r.Get("/available-articles/{id}", a.getAvailable)
		})
		r.With(
			a.hasRoleMiddleware("Spotlight PA"),
		).Group(func(r chi.Router) {
			r.Get("/upcoming-articles", a.listUpcoming)
			r.Post("/available-articles/{id}", a.postAvailable)
			r.Delete("/available-articles/{id}", a.deleteAvailable)
			r.Get("/message/{id}", a.getMessageFor)
			r.Post("/message", a.postMessage)
			r.Get("/scheduled-articles/{id}", a.getScheduledArticle)
			r.Post("/scheduled-articles", a.postScheduledArticle)
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
			if err := a.auth.HasRole(r, role); err != nil {
				a.errorResponse(w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (a *appEnv) listUpcoming(w http.ResponseWriter, r *http.Request) {
	a.Println("start listUpcoming")

	var feed arcjson.API
	if err := httpjson.Get(r.Context(), a.c, a.srcFeedURL, &feed); err != nil {
		a.errorResponse(w, err)
		return
	}

	arcsvc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	if err := arcsvc.StoreFeed(feed); err != nil {
		// Log failure but soldier on?
		a.Printf("DANGER: did not store feed: %v", err)
	}

	a.jsonResponse(http.StatusOK, w, feed)
}

func (a *appEnv) postAvailable(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	a.Printf("starting postAvailable %s", articleID)
	arcsvc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	if err := arcsvc.SetAvailablity(articleID, true); err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusAccepted, w, http.StatusText(http.StatusAccepted))
}

func (a *appEnv) deleteAvailable(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	a.Printf("starting deleteAvailable %s", articleID)
	arcsvc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	if err := arcsvc.SetAvailablity(articleID, false); err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusAccepted, w, http.StatusText(http.StatusAccepted))
}

func (a *appEnv) listAvailable(w http.ResponseWriter, r *http.Request) {
	a.Printf("starting listAvailable")
	arcsvc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	contents, err := arcsvc.GetAvailableFeed()
	if err != nil {
		a.errorResponse(w, err)
		return
	}

	a.jsonResponse(http.StatusOK, w, &arcjson.API{Contents: contents})
}

func (a *appEnv) getAvailable(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	a.Printf("starting getAvailable %s", articleID)

	arcsvc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	if err := arcsvc.IsAvailable(articleID); err != nil {
		a.errorResponse(w, err)
		return
	}
	// TODO: Could be simpler here. Maybe if I rewrite in Postgres
	contents, err := arcsvc.GetAvailableFeed()
	if err != nil {
		a.errorResponse(w, err)
		return
	}
	feed := arcjson.API{Contents: contents}
	article, err := feed.Get(articleID)
	if err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusOK, w, article)
}

func (a *appEnv) getMessageFor(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	a.Printf("starting getMessageFor %s", articleID)

	arcsvc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	feed, err := arcsvc.GetFeed()
	if err != nil {
		a.errorResponse(w, err)
		return
	}

	art, err := feed.Get(articleID)
	if err != nil {
		a.errorResponse(w, err)
		return
	}
	type response struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	var res response
	res.Subject, res.Body = art.Message()
	a.jsonResponse(http.StatusOK, w, &res)
}

func (a *appEnv) postMessage(w http.ResponseWriter, r *http.Request) {
	a.Printf("starting postMessage")
	type request struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	var req request
	if err := httpjson.DecodeRequest(w, r, &req); err != nil {
		a.errorResponse(w, err)
		return
	}
	if err := a.email.SendEmail(req.Subject, req.Body); err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusAccepted, w, http.StatusText(http.StatusAccepted))
}

func (a *appEnv) getScheduledArticle(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	a.Printf("start getScheduledArticle %s", articleID)

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

func (a *appEnv) postScheduledArticle(w http.ResponseWriter, r *http.Request) {
	a.Println("start postScheduledArticle")

	var userData almanack.ScheduledArticle
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		a.errorResponse(w, err)
		return
	}
	arcsvc := arcjson.FeedService{DataStore: a.store, Logger: a.Logger}
	sas := almanack.ScheduledArticleService{
		ArticleService: arcsvc,
		DataStore:      a.store,
		Logger:         a.Logger,
	}

	if err := sas.Save(&userData); err != nil {
		a.errorResponse(w, err)
		return
	}

	a.jsonResponse(http.StatusAccepted, w, &userData)
}

func (a *appEnv) getSignedUpload(w http.ResponseWriter, r *http.Request) {
	type response struct {
		SignedURL string `json:"signed-url"`
		FileName  string `json:"filename"`
	}
	var (
		res response
		err error
	)
	res.SignedURL, res.FileName, err = a.imageStore.GetSignedUpload()
	if err != nil {
		a.errorResponse(w, err)
		return
	}
	a.jsonResponse(http.StatusOK, w, &res)
}
