package httpx

import (
	"net/http"

	"github.com/jba/muxpatterns"
)

// TODO: replace with stdlib after https://github.com/golang/go/issues/61410
func PathValue(r *http.Request, name string) string {
	return muxpatterns.PathValue(r, name)
}
