package almlog

import (
	"errors"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
)

var HTTPTransport http.RoundTripper

func init() {
	HTTPTransport = requests.LogTransport(http.DefaultTransport, logReq)
	http.DefaultTransport = requests.ErrorTransport(errors.New("use of http.DefaultTransport"))
	http.DefaultClient.Transport = requests.ErrorTransport(errors.New("use of http.DefaultClient"))
}

func logReq(req *http.Request, res *http.Response, err error, duration time.Duration) {
	level := LevelThreshold(duration, 500*time.Millisecond, 1*time.Second)
	FromContext(req.Context()).
		Log(req.Context(), level, "RoundTrip",
			"req_method", req.Method,
			"req_host", req.Host,
			"res_status", res.StatusCode,
			"duration", duration,
		)
}
