package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func (app *appEnv) routes() http.Handler {
	r := chi.NewRouter()
	if app.isLambda {
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
	} else {
		r.Use(middleware.Recoverer)
	}
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.Logger}))
	r.Use(app.versionMiddleware)
	r.Use(app.maxSizeMiddleware)
	r.Get("/api/healthcheck", app.ping)
	r.Get(`/api/healthcheck/{code:\d{3}}`, app.pingErr)
	r.Get(`/api/proxy-image/{encURL}`, app.getProxyImage)
	r.Get(`/api/cron`, app.getCron)
	r.Get("/api/bookmarklet/{slug}", app.getBookmarklet)
	r.Route("/api", func(r chi.Router) {
		r.Use(app.authHeaderMiddleware)
		r.Get("/user-info", app.userInfo)
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
			r.Post("/create-signed-upload", app.postSignedUpload)
			r.Post("/image-update", app.postImageUpdate)
			r.Get("/authorized-domains", app.listDomains)
			r.Post("/authorized-domains", app.postDomain)
			r.Get("/authorized-addresses", app.listAddresses)
			r.Post("/authorized-addresses", app.postAddress)
			r.Get("/all-pages", app.listAllPages)
			r.Get(`/news-pages/{page:\d+}`, app.listNewsPages)
			r.Get(`/newsletter-pages/{page:\d+}`, app.listNewsletterPages)
			r.Get(`/page/{id:\d+}`, app.getPage)
			r.Get(`/page-with-content/{id:\d+}`, app.getPageWithContent)
			r.Post(`/page`, app.postPage)
			r.Get("/images", app.listImages)
			r.Get("/editors-picks", app.getEditorsPicks)
			r.Post("/editors-picks", app.postEditorsPicks)
			r.Get("/all-topics", app.listAllTopics)
			r.Get("/all-series", app.listAllSeries)
			r.Get("/files-list", app.listFiles)
			r.Post("/files-create", app.postFileCreate)
			r.Post("/files-update", app.postFileUpdate)
			r.Post("/page-for-arc-id", app.postPageForArcID)
		})
	})
	r.Route("/ssr", func(r chi.Router) {
		r.Use(app.authCookieMiddleware)
		// Don't trust this middleware!
		// Netlify should be verifying the role at the CDN level.
		// This is just a fallback.
		r.Use(app.hasRoleMiddleware("Spotlight PA"))
		r.Get("/user-info", app.userInfo)
		r.Get(`/page/{id:\d+}`, app.renderPage)
		r.NotFound(app.renderNotFound)
	})

	r.Route("/api-background", func(r chi.Router) {
		r.Get(`/sleep/{duration}`, app.backgroundSleep)
		r.Get(`/cron`, app.backgroundCron)
		r.Get(`/refresh-pages`, app.backgroundRefreshPages)
	})

	r.NotFound(app.notFound)

	return r
}
