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
	standardMW.Control(mux, `GET /api/arc-image`, app.getArcImage)
	standardMW.HandleFunc(mux, `GET /api/bookmarklet/{slug}`, app.getBookmarklet)
	standardMW.HandleFunc(mux, `GET /api/healthcheck`, app.ping)
	standardMW.HandleFunc(mux, `GET /api/healthcheck/{code}`, app.pingErr)
	standardMW.HandleFunc(mux, `POST /api/identity-hook`, app.postIdentityHook)
	// End public endpoints

	authMW := standardMW.With(app.authHeaderMiddleware)

	authMW.Control(mux, `GET /api/user-info`, app.userInfo)

	partnerMW := authMW.With(app.hasRoleMiddleware("editor"))

	// Start partner endpoints
	partnerMW.Control(mux, `GET /api/shared-article`, app.getSharedArticle)
	partnerMW.Control(mux, `GET /api/shared-articles`, app.listSharedArticles)
	// End partner endpoints

	spotlightMW := authMW.With(app.hasRoleMiddleware("Spotlight PA"))

	// Start Spotlight endpoints
	spotlightMW.HandleFunc(mux, `GET /api/all-series`, app.listAllSeries)
	spotlightMW.HandleFunc(mux, `GET /api/all-topics`, app.listAllTopics)
	spotlightMW.HandleFunc(mux, `GET /api/authorized-addresses`, app.listAddresses)
	spotlightMW.HandleFunc(mux, `POST /api/authorized-addresses`, app.postAddress)
	spotlightMW.HandleFunc(mux, `GET /api/authorized-domains`, app.listDomains)
	spotlightMW.HandleFunc(mux, `POST /api/authorized-domains`, app.postDomain)
	spotlightMW.HandleFunc(mux, `POST /api/create-signed-upload`, app.postSignedUpload)
	spotlightMW.Control(mux, `POST /api/donor-wall`, app.postDonorWall)
	spotlightMW.HandleFunc(mux, `POST /api/files-create`, app.postFileCreate)
	spotlightMW.HandleFunc(mux, `GET /api/files-list`, app.listFiles)
	spotlightMW.HandleFunc(mux, `POST /api/files-update`, app.postFileUpdate)
	spotlightMW.HandleFunc(mux, `GET /api/gdocs-doc`, app.getGDocsDoc)
	spotlightMW.HandleFunc(mux, `POST /api/gdocs-doc`, app.postGDocsDoc)
	spotlightMW.HandleFunc(mux, `POST /api/image-update`, app.postImageUpdate)
	spotlightMW.HandleFunc(mux, `GET /api/images`, app.listImages)
	spotlightMW.HandleFunc(mux, `POST /api/message`, app.postMessage)
	spotlightMW.HandleFunc(mux, `GET /api/page`, app.getPage)
	spotlightMW.HandleFunc(mux, `POST /api/page`, app.postPage)
	spotlightMW.HandleFunc(mux, `POST /api/page-create`, app.postPageCreate)
	spotlightMW.Control(mux, `POST /api/page-load`, app.postPageLoad)
	spotlightMW.HandleFunc(mux, `GET /api/pages`, app.listPages)
	spotlightMW.HandleFunc(mux, `GET /api/pages-by-fts`, app.listPagesByFTS)
	spotlightMW.HandleFunc(mux, `POST /api/page-refresh`, app.postPageRefresh)
	spotlightMW.HandleFunc(mux, `POST /api/shared-article`, app.postSharedArticle)
	spotlightMW.HandleFunc(mux, `POST /api/shared-article-from-gdocs`, app.postSharedArticleFromGDocs)
	spotlightMW.HandleFunc(mux, `GET /api/sidebar`, app.siteDataGet(almanack.SidebarLoc))
	spotlightMW.HandleFunc(mux, `POST /api/sidebar`, app.siteDataSet(almanack.SidebarLoc))
	spotlightMW.Control(mux, `GET /api/site-data`, app.getSiteData)
	spotlightMW.Control(mux, `POST /api/site-data`, app.postSiteData)
	spotlightMW.HandleFunc(mux, `GET /api/site-params`, app.siteDataGet(almanack.SiteParamsLoc))
	spotlightMW.HandleFunc(mux, `POST /api/site-params`, app.siteDataSet(almanack.SiteParamsLoc))
	// End spotlight endpoints

	// Don't trust this middleware!
	// Netlify should be verifying the role at the CDN level.
	// This is just a fallback.
	ssrMW := standardMW.With(app.authCookieMiddleware)

	ssrMW.Control(mux, `GET /ssr/user-info`, app.userInfo)
	ssrMW.HandleFunc(mux, `/ssr/`, app.renderNotFound)
	ssrMW.Control(mux, `GET /ssr/redirect/{slug}`, app.redirectSSR)

	partnerSSRMW := ssrMW.With(app.hasRoleMiddleware("editor"))

	partnerSSRMW.Control(mux, `GET /ssr/download-image`, app.redirectImageURL)

	// Start background API endpoints
	backgroundMW := mid.Stack{
		httpx.WithTimeout(14 * time.Minute),
	}
	backgroundMW.Control(mux, `GET /api-background/cron`, app.backgroundCron)
	backgroundMW.Control(mux, `GET /api-background/images`, app.backgroundImages)
	backgroundMW.Control(mux, `GET /api-background/refresh-pages`, app.backgroundRefreshPages)
	backgroundMW.Control(mux, `GET /api-background/sleep/{duration}`, app.backgroundSleep)
	// End background API endpoints

	standardMW.HandleFunc(mux, "/", app.notFound)

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
