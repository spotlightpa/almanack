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

func (app *appEnv) renderPage(w http.ResponseWriter, r *http.Request) {
	var id int64
	mustIntParam(r, "id", &id)
	app.logStart(r, "id", id)

	page, err := app.svc.Queries.GetPageByID(r.Context(), id)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find page ID %d", id)
		app.replyHTMLErr(w, r, err)
		return
	}
	title, _ := page.Frontmatter["title"].(string)
	body, _ := page.Frontmatter["raw-content"].(string)

	app.replyHTML(w, r, layouts.MailChimp, struct {
		Title string
		Body  template.HTML
	}{
		Title: title,
		Body:  template.HTML(body),
	})
}

func (app *appEnv) redirectImageURL(w http.ResponseWriter, r *http.Request) {
	src := r.URL.Query().Get("src")
	app.logStart(r, "src", src)
	if src == "" {
		http.Error(w, "Missing required parameter src", http.StatusBadRequest)
		return
	}
	redirect, err := app.svc.ImageStore.SignGetURL(r.Context(), src)
	if err != nil {
		app.logErr(r.Context(), err)
		app.replyHTMLErr(w, r, err)
		return
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}
