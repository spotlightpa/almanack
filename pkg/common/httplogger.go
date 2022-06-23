package common

import (
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
)

func init() {
	oldTransport := http.DefaultTransport
	http.DefaultTransport = requests.RoundTripFunc(func(req *http.Request) (res *http.Response, err error) {
		start := time.Now()
		res, err = oldTransport.RoundTrip(req)
		Logger.Printf("request to %s %q took %v",
			req.Method, req.Host, time.Since(start))
		return
	})
}
