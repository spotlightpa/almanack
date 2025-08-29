package api

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"maps"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/earthboundkid/resperr/v2"
	"github.com/getsentry/sentry-go"

	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/layouts"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (app *appEnv) replyJSON(statusCode int, w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	enc := json.NewEncoder(w)
	if err := enc.Encode(data); err != nil {
		// TODO: Log to Sentry
		almlog.Logger.Error("replyJSON", "err", err)
	}
}

func (app *appEnv) replyErr(w http.ResponseWriter, r *http.Request, err error) {
	app.logErr(r.Context(), err)
	code := resperr.StatusCode(err)
	details := url.Values{"": []string{resperr.UserMessage(err)}}
	maps.Insert(details, maps.All(resperr.ValidationErrors(err)))
	app.replyJSON(code, w, struct {
		Status  int        `json:"status"`
		Details url.Values `json:"details"`
	}{
		code,
		details,
	})
}

func (app *appEnv) replyNewErr(code int, w http.ResponseWriter, r *http.Request, format string, v ...any) {
	app.replyErr(w, r, resperr.New(code, format, v...))
}

func (app *appEnv) logErr(ctx context.Context, err error) {
	l := almlog.FromContext(ctx)
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			userinfo := netlifyid.FromContext(ctx)
			scope.SetTag("username", cmp.Or(userinfo.Username(), "anonymous"))
			scope.SetTag("email", cmp.Or(userinfo.Email(), "not set"))
			hub.CaptureException(err)
		})
	} else {
		l.Warn("sentry not in context")
	}
	l.ErrorContext(ctx, "logErr", "err", err)
}

func (app *appEnv) tryReadJSON(w http.ResponseWriter, r *http.Request, dst any) error {
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
		var (
			syntaxError        *json.SyntaxError
			unmarshalTypeError *json.UnmarshalTypeError
			maxBytesError      *http.MaxBytesError
		)

		switch {
		case errors.As(err, &syntaxError):
			return resperr.E{E: err, M: fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)}

		case errors.Is(err, io.ErrUnexpectedEOF):
			return resperr.E{E: err, M: "Request body contains badly-formed JSON"}

		case errors.As(err, &unmarshalTypeError):
			return resperr.E{E: err, M: fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return resperr.E{E: err, M: fmt.Sprintf("Request body contains unknown field %s", fieldName)}

		case errors.Is(err, io.EOF):
			return resperr.E{M: "Request body contains badly-formed JSON"}

		case errors.As(err, &maxBytesError):
			return resperr.New(http.StatusRequestEntityTooLarge,
				"request body exceeds max size %d: %w",
				maxBytesError.Limit, err)

		default:
			return resperr.New(http.StatusBadRequest, "tryReadJSON: %w", err)
		}
	}

	var discard any
	if err := dec.Decode(&discard); !errors.Is(err, io.EOF) {
		return resperr.E{M: "Request body must only contain a single JSON object"}
	}

	return nil
}

func (app *appEnv) readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
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
	const (
		megabyte = 1 << 20
		maxSize  = 5 * megabyte
	)
	return http.MaxBytesHandler(next, maxSize)
}

func intParam[Int int | int32 | int64](r *http.Request, param string, p *Int) error {
	pstr := r.PathValue(param)
	if pstr == "" {
		return fmt.Errorf("parameter %q not set", param)
	}
	if err := intFromString(pstr, p); err != nil {
		return (resperr.E{E: err, M: fmt.Sprintf("Bad integer parameter for %s", param)})
	}
	return nil
}

func intFromQuery[Int int | int32 | int64](r *http.Request, param string, p *Int) bool {
	s := r.URL.Query().Get(param)
	err := intFromString(s, p)
	return err == nil
}

func intFromString[Int int | int32 | int64](s string, p *Int) error {
	bitsize := 0
	switch any(p).(type) {
	case *int:
	case *int32:
		bitsize = 32
	case *int64:
		bitsize = 64
	default:
		panic("unreachable")
	}
	n, err := strconv.ParseInt(s, 10, bitsize)
	if err != nil {
		return err
	}
	*p = Int(n)
	return nil
}

func boolFromQuery(r *http.Request, param string) (val bool, err error) {
	if !r.URL.Query().Has(param) {
		return
	}
	s := r.URL.Query().Get(param)
	val, err = strconv.ParseBool(s)
	if err != nil {
		err = resperr.E{E: err, M: fmt.Sprintf("Could not interpret %s=%q", param, s)}
	}
	return
}

func (app *appEnv) replyHTML(w http.ResponseWriter, r *http.Request, t *template.Template, data any) {
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
	app.logErr(r.Context(), err)
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

func (app *appEnv) logStart(r *http.Request, args ...any) {
	pc, file, line, ok := runtime.Caller(1)
	route := "unknown"
	if ok {
		f := runtime.FuncForPC(pc)
		file = filepath.Base(file)
		_, name, _ := stringx.LastCut(f.Name(), ".")
		route = fmt.Sprintf("%s(%s:%d)", name, file, line)
	}
	l := almlog.FromContext(r.Context())
	l.With(args...).InfoContext(r.Context(), "logStart", "route", route)
}

func (app *appEnv) jsonErr(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.replyErr(w, r, err)
	})
}

func (app *appEnv) jsonNewErr(code int, format string, args ...any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.replyErr(w, r, resperr.New(code, format, args...))
	})
}

func (app *appEnv) jsonBadRequest(err error, format string, args ...any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.replyErr(w, r, resperr.E{E: err, M: fmt.Sprintf(format, args...)})
	})
}

func (app *appEnv) jsonOK(v any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.replyJSON(http.StatusOK, w, v)
	})
}

func (app *appEnv) jsonAccepted(v any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.replyJSON(http.StatusAccepted, w, v)
	})
}

func (app *appEnv) htmlOK(t *template.Template, data any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.replyHTML(w, r, t, data)
	})
}

func (app *appEnv) htmlErr(err error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.replyHTMLErr(w, r, err)
	})
}

func (app *appEnv) htmlBadRequest(err error, format string, args ...any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.replyHTMLErr(w, r, resperr.E{E: err, M: fmt.Sprintf(format, args...)})
	})
}
