package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"

	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/stringutils"
	"github.com/spotlightpa/almanack/layouts"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func (app *appEnv) replyJSON(statusCode int, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		app.Printf("replyJSON problem: %v", err)
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
		hub.WithScope(func(scope *sentry.Scope) {
			userinfo := netlifyid.FromContext(ctx)
			scope.SetTag("username", stringutils.First(userinfo.Username(), "anonymous"))
			scope.SetTag("email", stringutils.First(userinfo.Email(), "not set"))

			for _, suberr := range errutil.AsSlice(err) {
				hub.CaptureException(suberr)
			}
		})
	} else {
		app.Printf("sentry not in context")
	}
	app.Printf("err: %+v", err)
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

func (app *appEnv) authHeaderMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2, err := app.auth.AuthFromHeader(r)
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		h.ServeHTTP(w, r2)
	})
}

func (app *appEnv) authCookieMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2, err := app.auth.AuthFromCookie(r)
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		h.ServeHTTP(w, r2)
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

func (app *appEnv) intParam(r *http.Request, param string, v interface{}) error {
	pstr := chi.URLParam(r, param)
	if pstr == "" {
		return fmt.Errorf("parameter %q not set", param)
	}
	bitsize := 0
	switch v.(type) {
	case *int:
	case *int32:
		bitsize = 32
	case *int64:
		bitsize = 64
	default:
		return fmt.Errorf("unsupported type %T", v)
	}
	n, err := strconv.ParseInt(pstr, 10, bitsize)
	if err != nil {
		return resperr.New(
			http.StatusBadRequest,
			"bad integer parameter for %s: %w",
			param, err)
	}
	switch i := v.(type) {
	case *int:
		*i = int(n)
	case *int32:
		*i = int32(n)
	case *int64:
		*i = n
	}
	return nil
}

func (app *appEnv) FetchFeed(ctx context.Context) (*arc.API, error) {
	var feed arc.API
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

func (app *appEnv) replyHTML(w http.ResponseWriter, r *http.Request, t *template.Template, data interface{}) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		app.logErr(r.Context(), err)
		app.replyHTMLErr(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	if _, err := buf.WriteTo(w); err != nil {
		app.logErr(r.Context(), err)
		return
	}
}

func (app *appEnv) replyHTMLErr(w http.ResponseWriter, r *http.Request, err error) {
	code := resperr.StatusCode(err)
	var buf bytes.Buffer
	if err := layouts.Error.Execute(&buf, struct {
		Status     string
		StatusCode int
		Message    string
	}{
		Status:     http.StatusText(code),
		StatusCode: code,
		Message:    resperr.UserMessage(err),
	}); err != nil {
		app.logErr(r.Context(), err)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(code)
	if _, err := buf.WriteTo(w); err != nil {
		app.logErr(r.Context(), err)
		return
	}
}
