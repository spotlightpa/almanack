// Package jwthook does basic JWT checking for Netlify auth hooks
package jwthook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/resperr"
)

func Verify(secret, token string) *Claim {
	tokens := strings.Split(token, ".")
	if len(tokens) != 3 {
		return nil
	}
	encheader, encpayload, encsig := tokens[0], tokens[1], tokens[2]
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%s.%s", encheader, encpayload)
	sig, err := base64.RawURLEncoding.DecodeString(encsig)
	if err != nil {
		return nil
	}
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil
	}
	header, err := base64.RawURLEncoding.DecodeString(encheader)
	if err != nil {
		return nil
	}
	var c Claim
	if err := json.Unmarshal(header, &c.Header); err != nil {
		return nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(encpayload)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(payload, &c.Payload); err != nil {
		return nil
	}
	return &c
}

type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

var HMAC256 = Header{
	Algorithm: "HS256",
	Type:      "JWT",
}

type Payload struct {
	IssuedAt  int    `json:"iat,omitempty"`
	Expires   int    `json:"exp,omitempty"`
	NotBefore int    `json:"nbf,omitempty"`
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	SHA256    string `json:"sha256"`
}

var Window = 1 * time.Hour

type Claim struct {
	Header
	Payload
}

func (c *Claim) Validate(h Header, subject, issuer string, now time.Time) error {
	var v resperr.Validator
	v.AddIf("header", c.Header != h, "unexpected header")
	v.AddIf("subject", c.Subject != subject, "unexpected subject")
	v.AddIf("issuer", c.Issuer != issuer, "unexpected issuer")
	if c.IssuedAt != 0 {
		issued := time.Unix(int64(c.IssuedAt), 0)
		v.AddIf("iat", now.Before(issued), "in future: %v", c.IssuedAt)
		v.AddIfUnset("iat", now.Sub(issued) > Window, "too old: %v", c.IssuedAt)
	}
	notBefore := time.Unix(int64(c.NotBefore), 0)
	if c.NotBefore != 0 && now.Before(notBefore) {
		v.Add("nbf", "before nbf: %v", c.NotBefore)
	}
	expires := time.Unix(int64(c.NotBefore), 0)
	if c.Expires != 0 && now.After(expires) {
		v.Add("exp", "after exp: %v", c.Expires)
	}
	return v.Err()
}

func VerifyRequest(r *http.Request, secret, subject, issuer string, now time.Time, v any) error {
	token := r.Header.Get("X-Webhook-Signature")
	claim := Verify(secret, token)
	if claim == nil {
		return resperr.New(http.StatusUnauthorized, "could not verify request")
	}
	if err := claim.Validate(HMAC256, subject, issuer, now); err != nil {
		return err
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	h := sha256.New()
	h.Write(body)
	checksum := fmt.Sprintf("%x", h.Sum(nil))
	if checksum != claim.SHA256 {
		return fmt.Errorf("unexpected checksum %q != %q", checksum, claim.SHA256)
	}
	if err := json.Unmarshal(body, v); err != nil {
		return err
	}
	return nil
}
