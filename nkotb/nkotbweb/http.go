package nkotbweb

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/carlmjohnson/resperr"
	"github.com/carlmjohnson/rootdown"
	"github.com/getsentry/sentry-go"
	"github.com/spotlightpa/nkotb/blocko"
	"github.com/spotlightpa/nkotb/build"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (app *appEnv) routes() http.Handler {
	var rr rootdown.Router
	rr.Post("/api/auth-doc", app.authDoc, rootdown.RedirectFromSlash)
	rr.Get("/api/get-doc/*", app.getDoc, rootdown.RedirectFromSlash)
	if !app.isLambda() {
		rr.NotFound(http.FileServer(http.Dir("./public")).ServeHTTP)
	}
	return app.logRoute(&rr)
}

func (app *appEnv) logRoute(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("[%s] %q - %s", r.Method, r.URL.Path, r.RemoteAddr)
		h.ServeHTTP(w, r)
	})
}

func (app *appEnv) replyErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logErr(r.Context(), err)
	code := resperr.StatusCode(err)
	msg := resperr.UserMessage(err)
	http.Error(w, msg, code)
}

func (app *appEnv) logErr(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		logger.Printf("sentry not in context")
	}

	logger.Printf("err: %v", err)
}

func (app *appEnv) authDoc(w http.ResponseWriter, r *http.Request) {
	stateToken, err := makeStateToken()
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	docID := r.FormValue("docID")
	conf := &oauth2.Config{
		ClientID:     app.oauthClientID,
		ClientSecret: app.oauthClientSecret,
		RedirectURL:  app.buildURL(docID, stateToken),
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
	// Redirect user to Google's consent page to ask for permission
	url := conf.AuthCodeURL(stateToken)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *appEnv) buildURL(docID, state string) string {
	u := build.URL
	var data = struct {
		DocID, State string
	}{docID, state}
	b, _ := json.Marshal(&data)
	u.Path = "/api/get-doc/" + base64.URLEncoding.EncodeToString(b)
	return u.String()
}

func (app *appEnv) getDoc(w http.ResponseWriter, r *http.Request) {
	var b []byte
	if !rootdown.Get(r, "/api/get-doc/*", &b) {
		app.replyErr(w, r,
			resperr.New(http.StatusBadRequest, "could not parse request path"))
		return
	}
	var data struct {
		DocID, State string
	}
	if err := json.Unmarshal(b, &data); err != nil {
		app.replyErr(w, r, resperr.WithStatusCode(err, http.StatusBadRequest))
		return
	}
	stateToken := r.FormValue("state")
	if stateToken != data.State {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest,
			"token %q != %q",
			stateToken, data.State))
		return
	}
	code := r.FormValue("code")
	scope := r.FormValue("scope")
	conf := &oauth2.Config{
		ClientID:     app.oauthClientID,
		ClientSecret: app.oauthClientSecret,
		RedirectURL:  app.buildURL(data.DocID, data.State),
		Scopes:       strings.Split(scope, ","),
		Endpoint:     google.Endpoint,
	}
	tok, err := conf.Exchange(r.Context(), code)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	cl := conf.Client(r.Context(), tok)
	n, err := getDoc(r.Context(), cl, data.DocID)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	var buf bytes.Buffer
	html.Render(&buf, n)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	blocko.Blockerize(w, &buf)
}
