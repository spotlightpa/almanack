package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/getsentry/sentry-go"

	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

func (app *appEnv) versionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Almanack-App-Version", almanack.BuildVersion)
		h.ServeHTTP(w, r)
	})
}

func (app *appEnv) jsonResponse(statusCode int, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		app.Printf("jsonResponse problem: %v", err)
	}
}

func (app *appEnv) errorResponse(ctx context.Context, w http.ResponseWriter, err error) {
	app.logErr(ctx, err)
	errResp := errutil.ResponseFrom(err)
	app.jsonResponse(errResp.StatusCode, w, errResp)
}

func (app *appEnv) logErr(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	}

	app.Printf("err: %v", err)
}

func (app *appEnv) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := app.auth.AddToRequest(r)
		if err != nil {
			app.errorResponse(r.Context(), w, err)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (app *appEnv) hasRoleMiddleware(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := app.auth.HasRole(r, role); err != nil {
				app.errorResponse(r.Context(), w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
