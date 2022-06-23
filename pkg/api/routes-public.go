package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi/v5"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func (app *appEnv) notFound(w http.ResponseWriter, r *http.Request) {
	app.replyErr(w, r, resperr.NotFound(r))
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.Println("start ping")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	w.Write(b)
}

func (app *appEnv) pingErr(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	statusCode, _ := strconv.Atoi(code)
	app.Printf("start pingErr %q", code)

	app.replyErr(w, r, resperr.New(
		statusCode, "got test ping %q", code,
	))
}

var inkyURL = func() *url.URL {
	u, err := url.Parse("https://www.inquirer.com")
	if err != nil {
		panic(err)
	}
	return u
}()

func (app *appEnv) getProxyImage(w http.ResponseWriter, r *http.Request) {
	app.Println("start getProxyImage")

	encURL := chi.URLParam(r, "encURL")
	decURL, err := base64.URLEncoding.DecodeString(encURL)
	if err != nil {
		app.replyErr(w, r, resperr.WithUserMessagef(
			nil, "Could not decode URL param: %q", encURL,
		))
		return
	}
	u, err := inkyURL.Parse(string(decURL))
	if err != nil {
		app.replyErr(w, r, resperr.WithUserMessagef(
			nil, "Bad image URL: %q", decURL,
		))
		return
	}

	const urlWhitelist = `^https://[^/]*(.inquirer.com|.arcpublishing.com|arc-anglerfish-arc2-prod-pmn.s3.amazonaws.com)/`
	cl := *app.svc.Client
	cl.Transport = requests.PermitURLTransport(cl.Transport, urlWhitelist)
	body, ctype, err := almanack.FetchImageURL(r.Context(), &cl, u.String())
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	w.Header().Set("Content-Type", ctype)
	ext := strings.TrimPrefix(ctype, "image/")
	disposition := fmt.Sprintf(`attachment; filename="image.%s"`, ext)
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Cache-Control", "public, max-age=900")
	if _, err = w.Write(body); err != nil {
		app.logErr(r.Context(), err)
	}
}

func (app *appEnv) getBookmarklet(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	app.Printf("starting getBookmarklet for %q", slug)

	page, err := app.svc.Queries.GetPageByURLPath(r.Context(), "%"+slug+"%")
	if err != nil {
		if db.IsNotFound(err) {
			err = resperr.WithUserMessagef(
				db.NoRowsAs404(err, "could not find url-path %q", slug),
				"No matching page found for %s.", slug)
		}
		app.replyHTMLErr(w, r, err)
		return
	}
	http.Redirect(w, r,
		fmt.Sprintf("/admin/news/%d", page.ID),
		http.StatusTemporaryRedirect)
}
