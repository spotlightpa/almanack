package api

import (
	"html/template"
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/layouts"
)

func (app *appEnv) renderNotFound(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	app.replyHTMLErr(w, r, resperr.NotFound(r))
}

func (app *appEnv) renderPage(w http.ResponseWriter, r *http.Request) http.Handler {
	var id int64
	if err := intParam(r, "id", &id); err != nil {
		return app.htmlErr(err)
	}
	app.logStart(r, "id", id)

	page, err := app.svc.Queries.GetPageByID(r.Context(), id)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find page ID %d", id)
		return app.htmlErr(err)
	}
	title, _ := page.Frontmatter["title"].(string)
	body, _ := page.Frontmatter["raw-content"].(string)

	return app.htmlOK(layouts.MailChimp, struct {
		Title string
		Body  template.HTML
	}{
		Title: title,
		Body:  template.HTML(body),
	})
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
