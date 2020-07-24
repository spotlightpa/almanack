package almanack

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/carlmjohnson/crockford"
	"github.com/spotlightpa/almanack/pkg/common"
)

func GetSignedFileUpload(is common.FileStore, filename string) (signedURL, filepath string, err error) {
	filepath = makeFilePath(filename)
	var h http.Header
	h.Set("Content-Disposition", fmt.Sprintf(
		"attachment; filename*=UTF-8''%s",
		url.PathEscape(filename)))
	signedURL, err = is.GetSignedURL(filepath, h)
	return
}

func makeFilePath(filename string) string {
	var sb strings.Builder
	filename = slugify(filename)
	if filename == "" {
		filename = "-"
	}
	sb.Grow(len("uploads/12/34/1234/") + len(filename))
	sb.WriteString("uploads/")
	t := crockford.Time(crockford.Lower, time.Now())
	sb.Write(t[:2])
	sb.WriteString("/")
	sb.Write(t[:4])
	sb.WriteString("/")
	sb.Write(t[4:])
	sb.WriteString("/")
	sb.WriteString(filename)
	return sb.String()
}

func slugify(s string) string {
	hadDash := true
	f := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			hadDash = false
			return r - 'A' + 'a'
		case
			r >= 'a' && r <= 'z',
			r >= '0' && r <= '9',
			r == '.':
			hadDash = false
			return r
		case hadDash:
			return -1
		}
		hadDash = true
		return '-'
	}
	return strings.Map(f, s)
}
