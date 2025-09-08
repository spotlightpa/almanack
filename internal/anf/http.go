package anf

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/carlmjohnson/requests"
)

func HHMACSignRequest(req *http.Request, key, secret string, now time.Time) error {
	secretB, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("bad secret: %w", err)
	}
	h := hmac.New(sha256.New, secretB)

	h.Write([]byte(req.Method))
	h.Write([]byte(req.URL.String()))
	date := now.UTC().Format(time.RFC3339)
	h.Write([]byte(date))

	if req.Body != nil {
		// Read the body and replace with a new reader so it can be read again
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("reading request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(body))

		if len(body) > 0 {
			contentType := req.Header.Get("Content-Type")
			h.Write([]byte(contentType))
			h.Write(body)
		}
	}
	signature := h.Sum(nil)

	encodedSignature := base64.StdEncoding.EncodeToString(signature)

	authHeader := fmt.Sprintf(`HHMAC; key="%s"; signature="%s"; date="%s"`,
		key, encodedSignature, date)

	req.Header.Set("Authorization", authHeader)

	return nil
}

func HHMacTransport(key, secret string, rt requests.Transport) requests.Transport {
	return requests.RoundTripFunc(func(req *http.Request) (res *http.Response, err error) {
		r2 := *req
		r2.Header = r2.Header.Clone()
		if err := HHMACSignRequest(&r2, key, secret, time.Now()); err != nil {
			return nil, err
		}
		return rt.RoundTrip(&r2)
	})
}
