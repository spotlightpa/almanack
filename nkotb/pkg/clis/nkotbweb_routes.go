package clis

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/carlmjohnson/rootdown"
	"github.com/getsentry/sentry-go"
	"github.com/spotlightpa/nkotb/build"
	"github.com/spotlightpa/nkotb/pkg/blocko"
	"github.com/spotlightpa/nkotb/pkg/gdocs"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (app *nkotbWebAppEnv) routes() http.Handler {
	var rr rootdown.Router
	rr.Get("/api/convert", app.convert, rootdown.RedirectFromSlash)
	rr.Get("/api/auth-callback", app.authCallback, rootdown.RedirectFromSlash)
	if !app.isLambda() {
		rr.NotFound(http.FileServer(http.Dir("./public")).ServeHTTP)
	}
	return app.logRoute(&rr)
}

func (app *nkotbWebAppEnv) logRoute(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("[%s] %q - %s", r.Method, r.URL.Path, r.RemoteAddr)
		h.ServeHTTP(w, r)
	})
}

func (app *nkotbWebAppEnv) replyErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logErr(r.Context(), err)
	code := resperr.StatusCode(err)
	msg := resperr.UserMessage(err)
	http.Error(w, msg, code)
}

func (app *nkotbWebAppEnv) logErr(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		logger.Printf("sentry not in context")
	}

	logger.Printf("err: %v", err)
}

const (
	tokenCookie       = "google-token"
	stateCookie       = "google-state"
	redirectURLCookie = "google-redirect-url"
)

func (app *nkotbWebAppEnv) setCookie(w http.ResponseWriter, name string, v interface{}) {
	const oneMonth = 60 * 60 * 24 * 31
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		panic(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    base64.URLEncoding.EncodeToString(buf.Bytes()),
		MaxAge:   oneMonth,
		HttpOnly: true,
		Secure:   app.isLambda(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	mac := hmac.New(sha256.New, app.signingSecret)
	mac.Write(buf.Bytes())
	sig := mac.Sum(nil)

	http.SetCookie(w, &http.Cookie{
		Name:     name + "-signed",
		Value:    base64.URLEncoding.EncodeToString(sig),
		MaxAge:   oneMonth,
		HttpOnly: true,
		Secure:   app.isLambda(),
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

}

func (app *nkotbWebAppEnv) deleteCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   app.isLambda(),
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     name + "-signed",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   app.isLambda(),
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}

func (app *nkotbWebAppEnv) getCookie(r *http.Request, name string, v interface{}) bool {
	c, err := r.Cookie(name)
	if err != nil {
		return false
	}
	b, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil {
		return false
	}

	c, err = r.Cookie(name + "-signed")
	if err != nil {
		return false
	}
	cookieHMAC, err := base64.URLEncoding.DecodeString(c.Value)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, app.signingSecret)
	mac.Write(b)
	expectedMAC := mac.Sum(nil)
	if !hmac.Equal(cookieHMAC, expectedMAC) {
		return false
	}

	dec := gob.NewDecoder(bytes.NewReader(b))
	err = dec.Decode(v)
	return err == nil
}

func (app *nkotbWebAppEnv) googleConfig() *oauth2.Config {
	u := build.URL
	u.Path = "/api/auth-callback"
	return &oauth2.Config{
		ClientID:     app.oauthClientID,
		ClientSecret: app.oauthClientSecret,
		RedirectURL:  u.String(),
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
}

func (app *nkotbWebAppEnv) googleClient(r *http.Request) *http.Client {
	var tok oauth2.Token
	if !app.getCookie(r, tokenCookie, &tok) {
		return nil
	}
	if !tok.Valid() && tok.RefreshToken == "" {
		return nil
	}
	conf := app.googleConfig()
	return conf.Client(r.Context(), &tok)
}

func (app *nkotbWebAppEnv) convert(w http.ResponseWriter, r *http.Request) {
	docID := r.FormValue("docID")
	cl := app.googleClient(r)
	if cl == nil {
		app.authRedirect(w, r)
		return
	}
	doc, err := gdocs.Request(r.Context(), cl, docID)
	if err != nil {
		if requests.HasStatusErr(err, http.StatusNotFound) {
			msg := fmt.Sprintf("Document not found: %s", docID)
			http.Error(w, msg, http.StatusNotFound)
			return
		}
		if requests.HasStatusErr(err, http.StatusForbidden) {
			msg := fmt.Sprintf("You are not authorized to view %s", docID)
			http.Error(w, msg, http.StatusForbidden)
			return
		}
		app.deleteCookie(w, tokenCookie)
		app.replyErr(w, r, err)
		return
	}
	n := gdocs.Convert(doc)
	var buf bytes.Buffer
	html.Render(&buf, n)
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	blocko.HTMLToMarkdown(w, &buf)
}

func (app *nkotbWebAppEnv) authRedirect(w http.ResponseWriter, r *http.Request) {
	app.setCookie(w, redirectURLCookie, r.URL)

	stateToken, err := makeStateToken()
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.setCookie(w, stateCookie, stateToken)

	conf := app.googleConfig()
	// Redirect user to Google's consent page to ask for permission
	url := conf.AuthCodeURL(stateToken)
	w.Header().Set("Cache-Control", "no-cache")
	http.Redirect(w, r, url, http.StatusSeeOther)
}

func (app *nkotbWebAppEnv) authCallback(w http.ResponseWriter, r *http.Request) {
	var state string
	if !app.getCookie(r, stateCookie, &state) {
		app.replyErr(w, r, resperr.New(http.StatusUnauthorized, "no saved state"))
		return
	}
	app.deleteCookie(w, stateCookie)

	var redirect url.URL
	if !app.getCookie(r, redirectURLCookie, &redirect) {
		app.replyErr(w, r, resperr.New(http.StatusUnauthorized, "no redirect"))
		return
	}
	app.deleteCookie(w, redirectURLCookie)

	if callbackState := r.FormValue("state"); state != callbackState {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest,
			"token %q != %q",
			state, callbackState))
		return
	}
	conf := app.googleConfig()
	tok, err := conf.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.setCookie(w, tokenCookie, &tok)
	http.Redirect(w, r, redirect.String(), http.StatusSeeOther)
}
