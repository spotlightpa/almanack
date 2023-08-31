package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spotlightpa/almanack/internal/httpx"

	"github.com/carlmjohnson/be"
)

func TestMiddleware(t *testing.T) {
	mws := httpx.NewStack(
		func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("1"))
				h.ServeHTTP(w, r)
				w.Write([]byte("1"))
			})
		},
		func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("2"))
				h.ServeHTTP(w, r)
				w.Write([]byte("2"))
			})
		},
		func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("3"))
				h.ServeHTTP(w, r)
				w.Write([]byte("3"))
			})
		},
	)

	h := mws.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("h"))
	})

	// Be resiliant to mutation of the stack
	mws[0] = nil

	// Work once
	w := httptest.NewRecorder()
	h.ServeHTTP(w, nil)
	be.Equal(t, "123h321", w.Body.String())

	// Work multiple times
	w = httptest.NewRecorder()
	h.ServeHTTP(w, nil)
	be.Equal(t, "123h321", w.Body.String())
}
