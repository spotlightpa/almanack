// Package httpx contains http utilities.
package httpx

import (
	"mime"
)

func AttachmentName(filename string) string {
	return mime.FormatMediaType("attachment", map[string]string{
		"filename": filename,
	})
}
