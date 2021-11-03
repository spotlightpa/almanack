package build

import (
	_ "embed"
	"net/url"
	"strings"
)

//go:embed rev.txt
var Rev string

//go:embed url.txt
var embedurl string

var URL = func() url.URL {
	u, err := url.Parse(strings.TrimSpace(embedurl))
	if err != nil {
		panic(err)
	}
	return *u
}()
