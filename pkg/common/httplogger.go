package common

import (
	"io"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/go-chi/chi/v5/middleware"
)

func init() {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = requests.RoundTripFunc(func(req *http.Request) (res *http.Response, err error) {
		start := time.Now()
		res, err = oldTransport.RoundTrip(req)
		logReq(start, req)
		if err != nil {
			return
		}
		res.Body = closeLogger{start, req, res.Body}
		return
	})
}

func logReq(start time.Time, req *http.Request) {
	Logger.Printf("request to %s%s%s %s%q%s took %s%v%s",
		purple, req.Method, reset, cyan, req.Host, reset, yellow, time.Since(start), reset)
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

var (
	yellow = maybe("\033[33m")
	purple = maybe("\033[35;1m")
	cyan   = maybe("\033[36m")
	reset  = maybe("\033[0m")
)

func maybe(s string) string {
	if !middleware.IsTTY {
		return ""
	}
	return s
}
