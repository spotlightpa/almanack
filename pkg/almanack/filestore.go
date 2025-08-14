package almanack

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/crockford"
	"github.com/spotlightpa/almanack/internal/aws"
	"github.com/spotlightpa/almanack/internal/httpx"
	"github.com/spotlightpa/almanack/internal/stringx"
)

func GetSignedFileUpload(ctx context.Context, is aws.BlobStore, filename, mimetype string) (signedURL, fileURL, disposition, cachecontrol string, err error) {
	filepath := makeFilePath(filename)
	fileURL = is.BuildURL(filepath)
	h := http.Header{}
	disposition = httpx.AttachmentName(filename)
	h.Set("Content-Disposition", disposition)
	h.Set("Content-Type", mimetype)
	cachecontrol = "public,max-age=365000000,immutable"
	h.Set("Cache-Control", cachecontrol)
	signedURL, err = is.SignPutURL(ctx, filepath, h)
	return
}

func makeFilePath(filename string) string {
	var sb strings.Builder
	filename = stringx.SlugifyFilename(filename)
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

func UploadJSON(ctx context.Context, is aws.BlobStore, filepath, cachecontrol string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	h := make(http.Header, 3)
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", cachecontrol)
	return is.WriteFile(ctx, filepath, h, b)
}
