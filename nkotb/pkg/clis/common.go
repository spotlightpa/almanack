package clis

import (
	"crypto/rand"
	"encoding/base64"
)

func makeStateToken() (string, error) {
	var b [15]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b[:]), nil
}
