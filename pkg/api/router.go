package api

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jba/muxpatterns"

	"github.com/spotlightpa/almanack/internal/httpx"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (app *appEnv) routes() http.Handler {
	// TODO: replace with stdlib after https://github.com/golang/go/issues/61410
	mux := muxpatterns.NewServeMux()

	var baseMW httpx.Stack
	baseMW.Push(httpx.WithPathValue(mux))
	baseMW.Push(middleware.RealIP)
	baseMW.PushIf(!app.isLambda, middleware.Recoverer)
	baseMW.Push(
		almlog.Middleware,
		app.versionMiddleware,
		app.maxSizeMiddleware,
	)

	// Start public endpoints
	mux.Handle(`GET /api/arc-image`,
		baseMW.Controller(app.getArcImage))
	mux.Handle(`GET /api/bookmarklet/{slug}`,
		baseMW.HandlerFunc(app.getBookmarklet))
	mux.Handle(`GET /api/healthcheck`,
		baseMW.HandlerFunc(app.ping))
	mux.Handle(`GET /api/healthcheck/{code}`,
		baseMW.HandlerFunc(app.pingErr))
	mux.Handle(`POST /api/identity-hook`,
		baseMW.HandlerFunc(app.postIdentityHook))
	// End public endpoints

	authMW := baseMW.Clone()
	authMW.Push(app.authHeaderMiddleware)

	mux.Handle(`GET /api/user-info`,
		authMW.Controller(app.userInfo))

	partnerMW := authMW.Clone()
	partnerMW.Push(app.hasRoleMiddleware("editor"))

	// Start partner endpoints
	mux.Handle(`GET /api/mailchimp-signup-url`,
		partnerMW.HandlerFunc(app.getSignupURL)) // TODO: move to SSR
	mux.Handle(`GET /api/shared-article`,
		partnerMW.Controller(app.getSharedArticle))
	mux.Handle(`GET /api/shared-articles`,
		partnerMW.Controller(app.listSharedArticles))
	// End partner endpoints

	spotlightMW := authMW.Clone()
	spotlightMW.Push(app.hasRoleMiddleware("Spotlight PA"))

	// Start Spotlight endpoints
	mux.Handle(`GET /api/all-pages`,
		spotlightMW.HandlerFunc(app.listAllPages))
	mux.Handle(`GET /api/all-series`,
		spotlightMW.HandlerFunc(app.listAllSeries))
	mux.Handle(`GET /api/all-topics`,
		spotlightMW.HandlerFunc(app.listAllTopics))
	mux.Handle(`GET /api/arc-by-last-updated`,
		spotlightMW.HandlerFunc(app.listArcByLastUpdated))
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
	mux.Handle(`GET /api/editors-picks`,
		spotlightMW.HandlerFunc(app.getSiteData(almanack.HomepageLoc)))
	mux.Handle(`POST /api/editors-picks`,
		spotlightMW.HandlerFunc(app.setSiteData((almanack.HomepageLoc))))
	mux.Handle(`GET /api/election-feature`,
		spotlightMW.HandlerFunc(app.getSiteData(almanack.ElectionFeatLoc)))
	mux.Handle(`POST /api/election-feature`,
		spotlightMW.HandlerFunc(app.setSiteData((almanack.ElectionFeatLoc))))
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
	mux.Handle(`GET /api/pages`,
		spotlightMW.HandlerFunc(app.listPages))
	mux.Handle(`GET /api/pages-by-fts`,
		spotlightMW.HandlerFunc(app.listPagesByFTS))
	mux.Handle(`POST /api/page-refresh`,
		spotlightMW.HandlerFunc(app.postPageRefresh))
	mux.Handle(`POST /api/shared-article`,
		spotlightMW.HandlerFunc(app.postSharedArticle))
	mux.Handle(`POST /api/shared-article-from-arc`,
		spotlightMW.HandlerFunc(app.postSharedArticleFromArc))
	mux.Handle(`POST /api/shared-article-from-gdocs`,
		spotlightMW.HandlerFunc(app.postSharedArticleFromGDocs))
	mux.Handle(`GET /api/sidebar`,
		spotlightMW.HandlerFunc(app.getSiteData(almanack.SidebarLoc)))
	mux.Handle(`POST /api/sidebar`,
		spotlightMW.HandlerFunc(app.setSiteData((almanack.SidebarLoc))))
	mux.Handle(`GET /api/site-params`,
		spotlightMW.HandlerFunc(app.getSiteData(almanack.SiteParamsLoc)))
	mux.Handle(`POST /api/site-params`,
		spotlightMW.HandlerFunc(app.setSiteData((almanack.SiteParamsLoc))))
	mux.Handle(`GET /api/state-college-editor`,
		spotlightMW.HandlerFunc(app.getSiteData(almanack.StateCollegeLoc)))
	mux.Handle(`POST /api/state-college-editor`,
		spotlightMW.HandlerFunc(app.setSiteData((almanack.StateCollegeLoc))))
	// End spotlight endpoints

	ssrMW := baseMW.Clone()
	// Don't trust this middleware!
	// Netlify should be verifying the role at the CDN level.
	// This is just a fallback.
	ssrMW.Push(app.authCookieMiddleware)

	mux.Handle(`GET /ssr/user-info`,
		ssrMW.Controller(app.userInfo))
	mux.Handle(`/ssr/`,
		ssrMW.HandlerFunc(app.renderNotFound))

	partnerSSRMW := ssrMW.Clone()
	partnerSSRMW.Push(app.hasRoleMiddleware("editor"))

	mux.Handle(`GET /ssr/download-image`,
		partnerSSRMW.Controller(app.redirectImageURL))

	spotlightSSRMW := ssrMW.Clone()
	spotlightSSRMW.Push(app.hasRoleMiddleware("Spotlight PA"))

	mux.Handle(`GET /ssr/page/{id}`,
		spotlightSSRMW.Controller(app.renderPage))

	// Start background API endpoints
	mux.Handle(`GET /api-background/cron`,
		baseMW.Controller(app.backgroundCron))
	mux.Handle(`GET /api-background/images`,
		baseMW.Controller(app.backgroundImages))
	mux.Handle(`GET /api-background/refresh-pages`,
		baseMW.Controller(app.backgroundRefreshPages))
	mux.Handle(`GET /api-background/sleep/{duration}`,
		baseMW.Controller(app.backgroundSleep))
	// End background API endpoints

	mux.Handle("/", baseMW.HandlerFunc(app.notFound))
	return mux
}
