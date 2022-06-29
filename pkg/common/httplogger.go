package common

import (
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
		Logger.Printf("request to %s%s%s %s%q%s took %s%v%s",
			purple, req.Method, reset, cyan, req.Host, reset, yellow, time.Since(start), reset)
		return
	})
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
