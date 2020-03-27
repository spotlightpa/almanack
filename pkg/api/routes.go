package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/httpjson"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

func (app *appEnv) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.Logger}))
	r.Use(app.versionMiddleware)
	r.Get("/api/healthcheck", app.ping)
	r.Route("/api", func(r chi.Router) {
		r.Use(app.authMiddleware)
		r.Get("/user-info", app.userInfo)
		r.With(
			app.hasRoleMiddleware("editor"),
		).Group(func(r chi.Router) {
			r.Get("/available-articles", app.listAvailable)
			r.Get("/available-articles/{id}", app.getAvailable)
			r.Get("/mailchimp-signup-url", app.getSignupURL)
		})
		r.With(
			app.hasRoleMiddleware("Spotlight PA"),
		).Group(func(r chi.Router) {
			r.Get("/upcoming-articles", app.listUpcoming)
			r.Post("/available-articles", app.postAvailable)
			r.Post("/message", app.postMessage)
			r.Get("/scheduled-articles/{id}", app.getScheduledArticle)
			r.Post("/scheduled-articles", app.postScheduledArticle)
			r.Post("/get-signed-upload", app.getSignedUpload)
			r.Get("/authorized-domains", app.listDomains)
			r.Post("/authorized-domains", app.postDomain)
			r.Get("/spotlightpa-articles", app.listSpotlightPAArticles)
		})
	})
	r.NotFound(app.notFound)

	return r
}

func (app *appEnv) notFound(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(r.Context(), w, errutil.Response{
		StatusCode: http.StatusNotFound,
		Message:    http.StatusText(http.StatusNotFound),
		Cause:      fmt.Errorf("path not found: %s", r.URL.Path),
	})
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.Println("start ping")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	w.Write(b)
}

func (app *appEnv) userInfo(w http.ResponseWriter, r *http.Request) {
	app.Println("start userInfo")
	userinfo, err := netlifyid.FromRequest(r)
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, userinfo)
}

func (app *appEnv) listUpcoming(w http.ResponseWriter, r *http.Request) {
	app.Println("start listUpcoming")

	var (
		feed almanack.ArcAPI
		err  error
	)
	feed.Contents, err = app.svc.ListAllArticles(r.Context())
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, feed)
}

func (app *appEnv) postAvailable(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting postAvailable")

	var userData struct {
		ID         string          `json:"_id"`
		Note       string          `json:"almanack-note,omitempty"`
		Status     almanack.Status `json:"almanack-status,omitempty"`
		RefreshArc bool            `json:"almanack-refresh-arc"`
	}
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	var story almanack.ArcStory
	if userData.RefreshArc {
		var feed almanack.ArcAPI
		if err := httpjson.Get(r.Context(), app.c, app.srcFeedURL, &feed); err != nil {
			app.errorResponse(r.Context(), w, err)
			return
		}
		if err := app.svc.StoreFeed(r.Context(), &feed); err != nil {
			app.errorResponse(r.Context(), w, err)
			return
		}
		for i := range feed.Contents {
			if feed.Contents[i].ID == userData.ID {
				story = feed.Contents[i]
			}
		}
	}
	story.ID = userData.ID
	story.Note = userData.Note
	story.Status = userData.Status

	if err := app.svc.SaveAlmanackArticle(r.Context(), &story, userData.RefreshArc); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	app.jsonResponse(http.StatusAccepted, w, &userData)
}

func (app *appEnv) listAvailable(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listAvailable")
	type response struct {
		Contents []almanack.ArcStory `json:"contents"`
	}
	var (
		res response
		err error
	)
	if res.Contents, err = app.svc.GetAvailableFeed(r.Context()); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	app.jsonResponse(http.StatusOK, w, res)
}

func (app *appEnv) getAvailable(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	app.Printf("starting getAvailable %s", articleID)

	article, err := app.svc.GetArcStory(r.Context(), articleID)
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	if article.Status != almanack.StatusAvailable {
		// Let Spotlight PA users get article regardless of its status
		if err := app.auth.HasRole(r, "Spotlight PA"); err != nil {
			app.errorResponse(r.Context(), w, errutil.NotFound)
			return
		}
	}

	app.jsonResponse(http.StatusOK, w, article)
}

func (app *appEnv) postMessage(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting postMessage")
	type request struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	var req request
	if err := httpjson.DecodeRequest(w, r, &req); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	if err := app.email.SendEmail(req.Subject, req.Body); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	app.jsonResponse(http.StatusAccepted, w, http.StatusText(http.StatusAccepted))
}

func (app *appEnv) getScheduledArticle(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	app.Printf("start getScheduledArticle %s", articleID)

	article, err := app.svc.GetScheduledArticle(r.Context(), articleID)
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	app.jsonResponse(http.StatusOK, w, article)
}

func (app *appEnv) postScheduledArticle(w http.ResponseWriter, r *http.Request) {
	app.Println("start postScheduledArticle")

	var userData struct {
		almanack.SpotlightPAArticle
		RefreshArc bool `json:"almanack-refresh-arc"`
	}
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	if userData.RefreshArc {
		var feed almanack.ArcAPI
		if err := httpjson.Get(r.Context(), app.c, app.srcFeedURL, &feed); err != nil {
			app.errorResponse(r.Context(), w, err)
			return
		}
		if err := app.svc.StoreFeed(r.Context(), &feed); err != nil {
			app.errorResponse(r.Context(), w, err)
			return
		}
		if err := app.svc.ResetSpotlightPAArticleArcData(r.Context(), &userData.SpotlightPAArticle); err != nil {
			app.errorResponse(r.Context(), w, err)
			return
		}
	} else if strings.HasPrefix(userData.ImageURL, "http") {
		if imageurl, err := almanack.UploadFromURL(
			r.Context(), app.c, app.imageStore, userData.ImageURL,
		); err != nil {
			// Keep trucking, but don't publish
			app.logErr(r.Context(), fmt.Errorf("could not upload image: %w", err))
			userData.ImageURL = ""
			userData.ScheduleFor = nil
		} else {
			userData.ImageURL = imageurl
		}
	}

	if err := app.svc.SaveScheduledArticle(r.Context(), &userData.SpotlightPAArticle); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	app.jsonResponse(http.StatusAccepted, w, &userData.SpotlightPAArticle)
}

func (app *appEnv) getSignedUpload(w http.ResponseWriter, r *http.Request) {
	app.Printf("start getSignedUpload")
	type response struct {
		SignedURL string `json:"signed-url"`
		FileName  string `json:"filename"`
	}
	var (
		res response
		err error
	)
	res.SignedURL, res.FileName, err = almanack.GetSignedUpload(app.imageStore)
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, &res)
}

func (app *appEnv) getSignupURL(w http.ResponseWriter, r *http.Request) {
	app.Println("start getSignupURL")
	app.jsonResponse(http.StatusOK, w, app.mailchimpSignupURL)
}

func (app *appEnv) listDomains(w http.ResponseWriter, r *http.Request) {
	app.Println("start listDomains")
	type response struct {
		Domains []string `json:"domains"`
	}

	domains, err := app.svc.Querier.ListDomainsWithRole(r.Context(), "editor")
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, response{
		domains,
	})
}

func (app *appEnv) postDomain(w http.ResponseWriter, r *http.Request) {
	app.Println("start postDomain")
	type request struct {
		Domain string `json:"domain"`
	}
	type response struct {
		Domains []string `json:"domains"`
	}
	var req request
	if err := httpjson.DecodeRequest(w, r, &req); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	if _, err := app.svc.Querier.AppendRoleToDomain(
		r.Context(),
		db.AppendRoleToDomainParams{
			Domain: req.Domain,
			Role:   "editor",
		},
	); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	domains, err := app.svc.Querier.ListDomainsWithRole(r.Context(), "editor")
	if err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, response{
		domains,
	})
}

func (app *appEnv) listSpotlightPAArticles(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listSpotlightPAArticles")
	type response struct {
		Articles []db.ListSpotlightPAArticlesRow `json:"articles"`
	}
	var (
		res response
		err error
	)

	if res.Articles, err = app.svc.Querier.ListSpotlightPAArticles(r.Context()); err != nil {
		app.errorResponse(r.Context(), w, err)
		return
	}

	app.jsonResponse(http.StatusOK, w, res)
}
