package almlog

import (
	"io"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
	"golang.org/x/exp/slog"
)

func init() {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = requests.RoundTripFunc(func(req *http.Request) (res *http.Response, err error) {
		start := time.Now()
		res, err = oldTransport.RoundTrip(req)
		if err != nil {
			logReq(start, req)
			return
		}
		res.Body = closeLogger{start, req, res.Body}
		return
	})
}

func logReq(start time.Time, req *http.Request) {
	duration := time.Since(start)
	level := LevelThreshold(duration, 500*time.Millisecond, 1*time.Second)
	slog.FromContext(req.Context()).
		Log(level, "RoundTrip",
			"method", req.Method,
			"host", req.Host,
			"duration", duration)
}

type closeLogger struct {
	start time.Time
	req   *http.Request
	io.ReadCloser
}

func (cl closeLogger) Close() error {
	logReq(cl.start, cl.req)
	return cl.ReadCloser.Close()
}
