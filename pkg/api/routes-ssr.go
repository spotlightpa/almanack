package api

import (
	"net/http"

	"github.com/earthboundkid/resperr/v2"
	"github.com/spotlightpa/almanack/internal/db"
)

func (app *appEnv) renderNotFound(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	app.replyHTMLErr(w, r, resperr.NotFound(r))
}

func (app *appEnv) redirectImageURL(w http.ResponseWriter, r *http.Request) http.Handler {
	src := r.URL.Query().Get("src")
	app.logStart(r, "src", src)
	if src == "" {
		return app.htmlBadRequest(nil, "Missing required parameter src")
	}
	redirect, err := app.svc.ImageStore.SignGetURL(r.Context(), src)
	if err != nil {
		return app.htmlErr(err)
	}
	http.Redirect(w, r, redirect, http.StatusFound)
	return nil
}

func (app *appEnv) redirectSSR(w http.ResponseWriter, r *http.Request) http.Handler {
	app.logStart(r)
	from := r.URL.Path
	redirectInfo, err := app.svc.Queries.GetRedirect(r.Context(), from)
	if err != nil {
		return app.htmlErr(db.NoRowsAs404(err, "redirect not found: from=%q", from))
	}
	if err != nil {
		return app.htmlErr(err)
	}
	for _, role := range redirectInfo.Roles {

		if err := app.auth.HasRole(r, role); err == nil {
			http.Redirect(w, r, redirectInfo.To, int(redirectInfo.Code))
			return nil
		}
	}
	return app.htmlErr(resperr.New(http.StatusUnauthorized, "redirect missing role; want: %v", redirectInfo.Roles))
}
