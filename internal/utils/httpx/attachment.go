// Package httpx contains http utilities.
package httpx

import (
	"mime"
	"net/http"
)

func SetAttachmentName(h http.Header, filename string) {
	h.Set("Content-Disposition", AttachmentName(filename))
}

func AttachmentName(filename string) string {
	return mime.FormatMediaType("attachment", map[string]string{
		"filename": filename,
	})
}
