package httpx

import (
	"context"
	"net/http"
	"time"

	"github.com/earthboundkid/mid"
)

func WithTimeout(d time.Duration) mid.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, stop := context.WithTimeout(r.Context(), d)
			defer stop()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
