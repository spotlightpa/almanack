package httpx

import (
	"context"
	"net/http"

	"github.com/jba/muxpatterns"
)

type muxKey struct{}
type muxValue struct {
	*muxpatterns.ServeMux
	*http.Request
}

func WithPathValue(mux *muxpatterns.ServeMux) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, muxKey{}, muxValue{mux, r})
			r2 := r.WithContext(ctx)
			next.ServeHTTP(w, r2)
		})
	}
}

func PathValue(r *http.Request, name string) string {
	muxval, _ := r.Context().Value(muxKey{}).(muxValue)
	return muxval.PathValue(muxval.Request, name)
}
