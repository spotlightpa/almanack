package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"

	"github.com/spotlightpa/almanack/pkg/almanack"
)

func (app *appEnv) replyJSON(statusCode int, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		app.Printf("jsonResponse problem: %v", err)
	}
}

func (app *appEnv) replyErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logErr(r.Context(), err)
	code := resperr.StatusCode(err)
	msg := resperr.UserMessage(err)
	app.replyJSON(code, w, struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}{
		code,
		msg,
	})
}

func (app *appEnv) logErr(ctx context.Context, err error) {
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.CaptureException(err)
	} else {
		app.Printf("sentry not in context")
	}

	app.Printf("err: %v", err)
}

func (app *appEnv) tryReadJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Thanks to https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
	if ct := r.Header.Get("Content-Type"); ct != "" {
		value, _, _ := mime.ParseMediaType(ct)
		if value != "application/json" {
			return resperr.New(http.StatusUnsupportedMediaType,
				"request Content-Type must be application/json; got %s",
				ct)
		}
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return resperr.New(http.StatusBadRequest,
				"request body contains badly-formed JSON (at position %d): %w",
				syntaxError.Offset, err)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return resperr.New(http.StatusBadRequest,
				"request body contains badly-formed JSON: %w", err)

		case errors.As(err, &unmarshalTypeError):
			return resperr.New(http.StatusBadRequest,
				"request body contains an invalid value for the %q field (at position %d): %w",
				unmarshalTypeError.Field, unmarshalTypeError.Offset, err)

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return resperr.New(http.StatusBadRequest,
				"Request body contains unknown field %s: %w",
				fieldName, err)

		case errors.Is(err, io.EOF):
			return resperr.New(http.StatusBadRequest,
				"request body must not be empty")

		case err.Error() == "http: request body too large":
			return resperr.New(http.StatusRequestEntityTooLarge,
				"request body too large: %w", err)

		default:
			return resperr.New(http.StatusBadRequest, "unexpected error: %w", err)
		}
	}

	var discard interface{}
	if err := dec.Decode(&discard); !errors.Is(err, io.EOF) {
		return resperr.New(http.StatusBadRequest,
			"request body must only contain a single JSON object")
	}

	return nil
}

func (app *appEnv) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	if err := app.tryReadJSON(w, r, dst); err != nil {
		app.replyErr(w, r, err)
		return false
	}
	return true
}

func (app *appEnv) versionMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Almanack-App-Version", almanack.BuildVersion)
		h.ServeHTTP(w, r)
	})
}

func (app *appEnv) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r, err := app.auth.AddToRequest(r)
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (app *appEnv) hasRoleMiddleware(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := app.auth.HasRole(r, role); err != nil {
				app.replyErr(w, r, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (app *appEnv) maxSizeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const (
			megabyte = 1 << 20
			maxSize  = 5 * megabyte
		)
		r2 := *r // shallow copy
		r2.Body = http.MaxBytesReader(w, r.Body, maxSize)
		next.ServeHTTP(w, &r2)
	})
}

func (app *appEnv) getPage(r *http.Request, route string) (page int, err error) {
	if pageStr := chi.URLParam(r, "page"); pageStr != "" {
		if page, err = strconv.Atoi(pageStr); err != nil {
			err = resperr.New(http.StatusBadRequest,
				"bad argument to %s: %w", route, err)
			return
		}
	}
	return
}

func (app *appEnv) FetchFeed(ctx context.Context) (*almanack.ArcAPI, error) {
	var feed almanack.ArcAPI
	// Timeout needs to leave enough time to report errors to Sentry before
	// AWS kills the Lambda…
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	if err := requests.URL(app.srcFeedURL).
		Client(app.svc.Client).
		ToJSON(&feed).
		Fetch(ctx); err != nil {
		return nil, resperr.New(
			http.StatusBadGateway, "could not fetch Arc feed: %w", err)
	}
	return &feed, nil
}
