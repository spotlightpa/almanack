package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *appEnv) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.Logger}))
	r.Use(app.versionMiddleware)
	r.Get("/api/healthcheck", app.ping)
	r.Get(`/api/healthcheck/{code:\d{3}}`, app.pingErr)
	r.Get(`/api/proxy-image/{encURL}`, app.getProxyImage)
	r.Get(`/api/cron`, app.getCron)
	r.Route("/api", func(r chi.Router) {
		r.Use(app.authMiddleware)
		r.Get("/user-info", app.userInfo)
		r.Get("/bookmarklet/{slug}", app.getBookmarklet)
		r.With(
			app.hasRoleMiddleware("editor"),
		).Group(func(r chi.Router) {
			r.Get("/available-articles", app.listAvailableArcStories)
			r.Get(`/list-available/{page:\d+}`, app.listAvailableArcStories)
			r.Get("/available-articles/{id}", app.getArcStory)
			r.Get("/mailchimp-signup-url", app.getSignupURL)
		})
		r.With(
			app.hasRoleMiddleware("Spotlight PA"),
		).Group(func(r chi.Router) {
			r.Get("/upcoming-articles", app.listAllArcStories)
			r.Get(`/list-any-arc/{page:\d+}`, app.listAllArcStories)
			r.Get("/list-arc-refresh", app.listWithArcRefresh)
			r.Post("/available-articles", app.postAlmanackArcStory)
			r.Post("/message", app.postMessage)
			r.Get("/scheduled-articles/{id}", app.getScheduledArticle)
			r.Post("/scheduled-articles", app.postScheduledArticle)
			r.Post("/create-signed-upload", app.postSignedUpload)
			r.Post("/image-update", app.postImageUpdate)
			r.Get("/authorized-domains", app.listDomains)
			r.Post("/authorized-domains", app.postDomain)
			r.Get("/spotlightpa-articles", app.listSpotlightPAArticles)
			r.Get("/images", app.listImages)
			r.Get("/editors-picks", app.getEditorsPicks)
			r.Post("/editors-picks", app.postEditorsPicks)
			r.Get("/all-topics", app.listAllTopics)
			r.Get("/all-series", app.listAllSeries)
			r.Get("/files-list", app.listFiles)
			r.Post("/files-create", app.postFileCreate)
			r.Post("/files-update", app.postFileUpdate)
		})
	})
	r.NotFound(app.notFound)

	return r
}
