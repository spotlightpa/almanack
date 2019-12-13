package api

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
)

type requestCache struct {
	m sync.Map
	r http.RoundTripper
	*log.Logger
}

func SetRounderTripper(c *http.Client, l *log.Logger) {
	r := c.Transport
	if r == nil {
		r = http.DefaultTransport
	}
	c.Transport = &requestCache{
		r:      r,
		Logger: l,
	}
}

func (rc *requestCache) Get(req *http.Request) (*http.Response, bool) {
	key := [...]string{req.Method, req.URL.String()}
	v, ok := rc.m.Load(key)
	if !ok {
		return nil, false
	}
	b := v.([]byte)
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(b)), req)
	if err != nil {
		rc.Printf("unexpected cache get error: %v", err)
		return nil, false
	}
	return resp, true
}

func (rc *requestCache) Set(req *http.Request, resp *http.Response) error {
	b, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return err
	}
	key := [...]string{req.Method, req.URL.String()}
	rc.m.Store(key, b)
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
