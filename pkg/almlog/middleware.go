package almlog

import (
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/go-chi/chi/v5/middleware"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		l := Logger
		// Add Lambda Request ID if present
		if lctx, ok := lambdacontext.FromContext(r.Context()); ok {
			reqID := lctx.AwsRequestID
			// Just first 8 chars
			reqID = reqID[:min(8, len(reqID))]
			l = l.With("req_id", reqID)
		}
		ctx := NewContext(r.Context(), l)
		r = r.WithContext(ctx)

		defer func() {
			duration := time.Since(start)
			level := LevelThreshold(duration, 1*time.Second, 5*time.Second)
			status := ww.Status()
			if l2 := LevelThreshold(status, 400, 500); l2 > level {
				level = l2
			}
			l.Log(r.Context(), level, "ServeHTTP",
				"req_method", r.Method,
				"req_ip", r.RemoteAddr,
				"req_path", r.RequestURI,
				"req_agent", r.UserAgent(),
				"res_status", status,
				"res_size", ww.BytesWritten(),
				"res_content_type", ww.Header().Get("Content-Type"),
				"duration", duration,
			)
		}()
		next.ServeHTTP(ww, r)
	})
}
