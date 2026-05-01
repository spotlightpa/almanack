package almapp

import (
	"net/http"
	"time"

	"github.com/earthboundkid/mid"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/almsvc"
	"github.com/spotlightpa/almanack/internal/utils/httpx"
)

func (app *appEnv) routes() http.Handler {
	mux := http.NewServeMux()

	standardMW := mid.Stack{
		httpx.WithTimeout(9 * time.Second),
	}

	// Start public endpoints
	standardMW.
		Control(mux, `GET /api/arc-image`, app.getArcImage).
		HandleFunc(mux, `GET /api/bookmarklet/{slug}`, app.getBookmarklet).
		HandleFunc(mux, `GET /api/healthcheck`, app.ping).
		HandleFunc(mux, `GET /api/healthcheck/{code}`, app.pingErr).
		HandleFunc(mux, `POST /api/identity-hook`, app.postIdentityHook)
	// End public endpoints

	authMW := standardMW.With(app.authHeaderMiddleware)

	authMW.Control(mux, `GET /api/user-info`, app.userInfo)

	partnerMW := authMW.With(app.hasRoleMiddleware("editor"))

	// Start partner endpoints
	partnerMW.
		Control(mux, `GET /api/shared-article`, app.getSharedArticle).
		Control(mux, `GET /api/shared-articles`, app.listSharedArticles)
	// End partner endpoints

	spotlightMW := authMW.With(app.hasRoleMiddleware("Spotlight PA"))

	// Start Spotlight endpoints
	spotlightMW.
		HandleFunc(mux, `GET /api/all-series`, app.listAllSeries).
		HandleFunc(mux, `GET /api/all-topics`, app.listAllTopics).
		HandleFunc(mux, `GET /api/authorized-addresses`, app.listAddresses).
		HandleFunc(mux, `POST /api/authorized-addresses`, app.postAddress).
		HandleFunc(mux, `GET /api/authorized-domains`, app.listDomains).
		HandleFunc(mux, `POST /api/authorized-domains`, app.postDomain).
		HandleFunc(mux, `POST /api/create-signed-upload`, app.postSignedUpload).
		Control(mux, `POST /api/donor-wall`, app.postDonorWall).
		HandleFunc(mux, `POST /api/files-create`, app.postFileCreate).
		HandleFunc(mux, `GET /api/files-list`, app.listFiles).
		HandleFunc(mux, `POST /api/files-update`, app.postFileUpdate).
		HandleFunc(mux, `GET /api/gdocs-doc`, app.getGDocsDoc).
		HandleFunc(mux, `POST /api/gdocs-doc`, app.postGDocsDoc).
		HandleFunc(mux, `POST /api/image-update`, app.postImageUpdate).
		HandleFunc(mux, `GET /api/images`, app.listImages).
		HandleFunc(mux, `POST /api/message`, app.postMessage).
		HandleFunc(mux, `GET /api/page`, app.getPage).
		HandleFunc(mux, `POST /api/page`, app.postPage).
		HandleFunc(mux, `POST /api/page-json`, app.postPageJSON).
		HandleFunc(mux, `POST /api/page-create`, app.postPageCreate).
		Control(mux, `POST /api/page-load`, app.postPageLoad).
		HandleFunc(mux, `GET /api/pages`, app.listPages).
		HandleFunc(mux, `GET /api/pages-by-fts`, app.listPagesByFTS).
		HandleFunc(mux, `POST /api/page-refresh`, app.postPageRefresh).
		HandleFunc(mux, `POST /api/shared-article`, app.postSharedArticle).
		HandleFunc(mux, `POST /api/shared-article-from-gdocs`, app.postSharedArticleFromGDocs).
		HandleFunc(mux, `GET /api/sidebar`, app.siteDataGet(almsvc.SidebarLoc)).
		HandleFunc(mux, `POST /api/sidebar`, app.siteDataSet(almsvc.SidebarLoc)).
		Control(mux, `GET /api/site-data`, app.getSiteData).
		Control(mux, `POST /api/site-data`, app.postSiteData).
		HandleFunc(mux, `GET /api/site-params`, app.siteDataGet(almsvc.SiteParamsLoc)).
		HandleFunc(mux, `POST /api/site-params`, app.siteDataSet(almsvc.SiteParamsLoc))
	// End spotlight endpoints

	// Don't trust this middleware!
	// Netlify should be verifying the role at the CDN level.
	// This is just a fallback.
	ssrMW := standardMW.With(app.authCookieMiddleware)

	ssrMW.
		Control(mux, `GET /ssr/user-info`, app.userInfo).
		HandleFunc(mux, `/ssr/`, app.renderNotFound).
		Control(mux, `GET /ssr/redirect/{slug}`, app.redirectSSR)

	partnerSSRMW := ssrMW.With(app.hasRoleMiddleware("editor"))

	partnerSSRMW.
		Control(mux, `GET /ssr/download-image`, app.redirectImageURL)

	// Start background API endpoints
	backgroundMW := mid.Stack{
		httpx.WithTimeout(14 * time.Minute),
	}
	backgroundMW.
		Control(mux, `GET /api-background/cron`, app.backgroundCron).
		Control(mux, `GET /api-background/images`, app.backgroundImages).
		Control(mux, `GET /api-background/refresh-pages`, app.backgroundRefreshPages).
		Control(mux, `GET /api-background/sleep/{duration}`, app.backgroundSleep)
	// End background API endpoints

	standardMW.HandleFunc(mux, "/", app.notFound)

	var baseMW mid.Stack

	corsProtection := http.NewCrossOriginProtection()
	corsProtection.SetDenyHandler(app.badCORS(corsProtection))

	baseMW.Push(
		middleware.RealIP,
		almlog.Middleware,
	)
	baseMW.PushIf(!app.isLambda, middleware.Recoverer)
	baseMW.PushIf(app.isLambda, sentryhttp.
		New(sentryhttp.Options{
			WaitForDelivery: true,
			Timeout:         5 * time.Second,
			Repanic:         false,
		}).Handle)
	baseMW.Push(
		corsProtection.Handler,
		app.versionMiddleware,
		app.maxSizeMiddleware,
	)
	return baseMW.Handler(mux)
}
