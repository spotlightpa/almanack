// Package httpx contains http utilities.
package httpx

import (
	"fmt"
	"net/http"
	"net/url"
)

func SetAttachmentName(h http.Header, filename string) {
	h.Set("Content-Disposition", AttachmentName(filename))
}

func AttachmentName(filename string) string {
	return fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(filename))
}
