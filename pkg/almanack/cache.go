package almanack

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
)

type requestCache struct {
	m map[[2]string][]byte
	l sync.RWMutex
	r http.RoundTripper
	*log.Logger
}

func SetRounderTripper(c *http.Client, l *log.Logger) {
	r := c.Transport
	if r == nil {
		r = http.DefaultTransport
	}
	c.Transport = &requestCache{
		m:      make(map[[2]string][]byte),
		r:      r,
		Logger: l,
	}
}

func (rc *requestCache) Get(req *http.Request) (*http.Response, bool) {
	rc.l.RLock()
	defer rc.l.RUnlock()

	b, ok := rc.m[[...]string{req.Method, req.URL.String()}]
	if !ok {
		return nil, false
	}
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(b)), req)
	if err != nil {
		rc.Printf("unexpected cache get error: %v", err)
		return nil, false
	}
	return resp, true
}

func (rc *requestCache) Set(req *http.Request, resp *http.Response) error {
	rc.l.Lock()
	defer rc.l.Unlock()

	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	rc.m[[...]string{req.Method, req.URL.String()}] = b
	fullresp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(b)), req)
	*resp = *fullresp
	return err
}

func (rc *requestCache) RoundTrip(req *http.Request) (*http.Response, error) {
	cacheable := req.Method == http.MethodGet
	if cacheable {
		if resp, ok := rc.Get(req); ok {
			rc.Printf("cache hit for %s", req.URL.String())
			return resp, nil
		}
		rc.Printf("cache miss for %s", req.URL.String())
	}
	resp, err := rc.r.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if cacheable {
		err = rc.Set(req, resp)
	}
	return resp, err
}
