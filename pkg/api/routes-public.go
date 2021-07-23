package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi"
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
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest, "could not decode URL param: %w", err,
		))
		return
	}
	u, err := inkyURL.Parse(string(decURL))
	if err != nil {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest, "bad image URL: %s", decURL,
		))
		return
	}
	app.Printf("requested %q", u)
	inWhitelist := false
	for _, domain := range []string{
		".inquirer.com",
		".arcpublishing.com",
	} {
		if strings.HasSuffix(u.Host, domain) {
			inWhitelist = true
			break
		}
	}
	if u.Host == "arc-anglerfish-arc2-prod-pmn.s3.amazonaws.com" {
		inWhitelist = true
	}
	if !inWhitelist {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest, "untrusted image URL: %s", u,
		))
		return
	}
	body, ctype, err := almanack.FetchImageURL(r.Context(), app.svc.Client, u.String())
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

func (app *appEnv) getCron(w http.ResponseWriter, r *http.Request) {
	if err := errutil.ExecParallel(func() error {
		return app.svc.PopScheduledArticles(r.Context())
	}, func() error {
		return app.svc.PopScheduledPages(r.Context())
	}, func() error {
		return app.svc.UpdateMostPopular(r.Context())
	}); err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, "OK")
}

func (app *appEnv) getBookmarklet(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	app.Printf("starting getBookmarklet for %q", slug)

	arcid, err := app.svc.Queries.GetArticleIDFromSlug(r.Context(), slug)
	if err != nil && !db.IsNotFound(err) {
		app.logErr(r.Context(), err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	if arcid == "" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r,
		fmt.Sprintf("/articles/%s/schedule", arcid),
		http.StatusTemporaryRedirect)
}
