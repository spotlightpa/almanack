package clis

import (
	"crypto/rand"
	"encoding/base64"
	"runtime/debug"
)

func makeStateToken() (string, error) {
	var b [15]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b[:]), nil
}

func getVersion() string {
	if i, ok := debug.ReadBuildInfo(); ok {
		return i.Main.Version
	}

	return "(unknown)"
}
