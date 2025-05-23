package api

import (
	"net/http"
	"time"

	"github.com/earthboundkid/mid"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/spotlightpa/almanack/internal/httpx"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (app *appEnv) routes() http.Handler {
	mux := http.NewServeMux()

	standardMW := mid.Stack{
		httpx.WithTimeout(9 * time.Second),
	}

	// Start public endpoints
	mux.Handle(`GET /api/arc-image`,
		standardMW.Controller(app.getArcImage))
	mux.Handle(`GET /api/bookmarklet/{slug}`,
		standardMW.HandlerFunc(app.getBookmarklet))
	mux.Handle(`GET /api/healthcheck`,
		standardMW.HandlerFunc(app.ping))
	mux.Handle(`GET /api/healthcheck/{code}`,
		standardMW.HandlerFunc(app.pingErr))
	mux.Handle(`POST /api/identity-hook`,
		standardMW.HandlerFunc(app.postIdentityHook))
	// End public endpoints

	authMW := standardMW.With(app.authHeaderMiddleware)

	mux.Handle(`GET /api/user-info`,
		authMW.Controller(app.userInfo))

	partnerMW := authMW.With(app.hasRoleMiddleware("editor"))

	// Start partner endpoints
	mux.Handle(`GET /api/shared-article`,
		partnerMW.Controller(app.getSharedArticle))
	mux.Handle(`GET /api/shared-articles`,
		partnerMW.Controller(app.listSharedArticles))
	// End partner endpoints

	spotlightMW := authMW.With(app.hasRoleMiddleware("Spotlight PA"))

	// Start Spotlight endpoints
	mux.Handle(`GET /api/all-series`,
		spotlightMW.HandlerFunc(app.listAllSeries))
	mux.Handle(`GET /api/all-topics`,
		spotlightMW.HandlerFunc(app.listAllTopics))
	mux.Handle(`GET /api/authorized-addresses`,
		spotlightMW.HandlerFunc(app.listAddresses))
	mux.Handle(`POST /api/authorized-addresses`,
		spotlightMW.HandlerFunc(app.postAddress))
	mux.Handle(`GET /api/authorized-domains`,
		spotlightMW.HandlerFunc(app.listDomains))
	mux.Handle(`POST /api/authorized-domains`,
		spotlightMW.HandlerFunc(app.postDomain))
	mux.Handle(`POST /api/create-signed-upload`,
		spotlightMW.HandlerFunc(app.postSignedUpload))
	mux.Handle(`POST /api/donor-wall`,
		spotlightMW.Controller(app.postDonorWall))
	mux.Handle(`POST /api/files-create`,
		spotlightMW.HandlerFunc(app.postFileCreate))
	mux.Handle(`GET /api/files-list`,
		spotlightMW.HandlerFunc(app.listFiles))
	mux.Handle(`POST /api/files-update`,
		spotlightMW.HandlerFunc(app.postFileUpdate))
	mux.Handle(`GET /api/gdocs-doc`,
		spotlightMW.HandlerFunc(app.getGDocsDoc))
	mux.Handle(`POST /api/gdocs-doc`,
		spotlightMW.HandlerFunc(app.postGDocsDoc))
	mux.Handle(`POST /api/image-update`,
		spotlightMW.HandlerFunc(app.postImageUpdate))
	mux.Handle(`GET /api/images`,
		spotlightMW.HandlerFunc(app.listImages))
	mux.Handle(`POST /api/message`,
		spotlightMW.HandlerFunc(app.postMessage))
	mux.Handle(`GET /api/page`,
		spotlightMW.HandlerFunc(app.getPage))
	mux.Handle(`POST /api/page`,
		spotlightMW.HandlerFunc(app.postPage))
	mux.Handle(`POST /api/page-create`,
		spotlightMW.HandlerFunc(app.postPageCreate))
	mux.Handle(`POST /api/page-load`,
		spotlightMW.Controller(app.postPageLoad))
	mux.Handle(`GET /api/pages`,
		spotlightMW.HandlerFunc(app.listPages))
	mux.Handle(`GET /api/pages-by-fts`,
		spotlightMW.HandlerFunc(app.listPagesByFTS))
	mux.Handle(`POST /api/page-refresh`,
		spotlightMW.HandlerFunc(app.postPageRefresh))
	mux.Handle(`POST /api/shared-article`,
		spotlightMW.HandlerFunc(app.postSharedArticle))
	mux.Handle(`POST /api/shared-article-from-gdocs`,
		spotlightMW.HandlerFunc(app.postSharedArticleFromGDocs))
	mux.Handle(`GET /api/sidebar`,
		spotlightMW.HandlerFunc(app.siteDataGet(almanack.SidebarLoc)))
	mux.Handle(`POST /api/sidebar`,
		spotlightMW.HandlerFunc(app.siteDataSet((almanack.SidebarLoc))))
	mux.Handle(`GET /api/site-data`,
		spotlightMW.Controller(app.getSiteData))
	mux.Handle(`POST /api/site-data`,
		spotlightMW.Controller(app.postSiteData))
	mux.Handle(`GET /api/site-params`,
		spotlightMW.HandlerFunc(app.siteDataGet(almanack.SiteParamsLoc)))
	mux.Handle(`POST /api/site-params`,
		spotlightMW.HandlerFunc(app.siteDataSet((almanack.SiteParamsLoc))))
	// End spotlight endpoints

	// Don't trust this middleware!
	// Netlify should be verifying the role at the CDN level.
	// This is just a fallback.
	ssrMW := standardMW.With(app.authCookieMiddleware)

	mux.Handle(`GET /ssr/user-info`,
		ssrMW.Controller(app.userInfo))
	mux.Handle(`/ssr/`,
		ssrMW.HandlerFunc(app.renderNotFound))

	partnerSSRMW := ssrMW.With(app.hasRoleMiddleware("editor"))

	mux.Handle(`GET /ssr/download-image`,
		partnerSSRMW.Controller(app.redirectImageURL))
	mux.Handle(`GET /ssr/mailchimp-signup-url`,
		partnerSSRMW.HandlerFunc(app.redirectSignupURL))

	spotlightSSRMW := ssrMW.With(app.hasRoleMiddleware("Spotlight PA"))

	mux.Handle(`GET /ssr/donor-wall`,
		spotlightSSRMW.Controller(app.redirectDonorWall))
	mux.Handle(`GET /ssr/page/{id}`,
		spotlightSSRMW.Controller(app.renderPage))

	// Start background API endpoints
	backgroundMW := mid.Stack{
		httpx.WithTimeout(14 * time.Minute),
	}
	mux.Handle(`GET /api-background/cron`,
		backgroundMW.Controller(app.backgroundCron))
	mux.Handle(`GET /api-background/images`,
		backgroundMW.Controller(app.backgroundImages))
	mux.Handle(`GET /api-background/refresh-pages`,
		backgroundMW.Controller(app.backgroundRefreshPages))
	mux.Handle(`GET /api-background/sleep/{duration}`,
		backgroundMW.Controller(app.backgroundSleep))
	// End background API endpoints

	mux.Handle("/", standardMW.HandlerFunc(app.notFound))

	var baseMW mid.Stack
	baseMW.Push(middleware.RealIP)
	baseMW.PushIf(!app.isLambda, middleware.Recoverer)
	baseMW.Push(
		almlog.Middleware,
		app.versionMiddleware,
		app.maxSizeMiddleware,
	)
	return baseMW.Handler(mux)
}
