package almanack

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/carlmjohnson/crockford"
	"github.com/spotlightpa/almanack/internal/aws"
)

func GetSignedFileUpload(ctx context.Context, is aws.BlobStore, filename, mimetype string) (signedURL, fileURL, disposition, cachecontrol string, err error) {
	filepath := makeFilePath(filename)
	fileURL = is.BuildURL(filepath)
	h := http.Header{}
	disposition = fmt.Sprintf("attachment; filename*=UTF-8''%s",
		url.PathEscape(filename))
	h.Set("Content-Disposition", disposition)
	h.Set("Content-Type", mimetype)
	cachecontrol = "public,max-age=365000000,immutable"
	h.Set("Cache-Control", cachecontrol)
	signedURL, err = is.GetSignedURL(ctx, filepath, h)
	return
}

func makeFilePath(filename string) string {
	var sb strings.Builder
	filename = slugify(filename)
	if filename == "" {
		filename = "-"
	}
	sb.Grow(len("uploads/1234/1234/") + len(filename))
	sb.WriteString("uploads/")
	t := crockford.Time(crockford.Lower, time.Now())
	sb.WriteString(t[:4])
	sb.WriteString("/")
	sb.WriteString(t[4:])
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

func UploadJSON(ctx context.Context, is aws.BlobStore, filepath, cachecontrol string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	h := make(http.Header, 2)
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", cachecontrol)
	return is.WriteFile(ctx, filepath, h, b)
}
