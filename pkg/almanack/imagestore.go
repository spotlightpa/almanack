package almanack

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/resperr"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spotlightpa/almanack/internal/aws"
	"golang.org/x/net/context/ctxhttp"
)

func GetSignedImageUpload(ctx context.Context, is aws.BlobStore, ct string) (signedURL, filename string, err error) {
	var ext string
	if i := strings.Index(ct, "/"); i == -1 || i+1 >= len(ct) {
		return "", "", resperr.New(
			http.StatusBadRequest, "bad mimetype %q", ct)
	} else {
		ext = ct[i:]
	}
	filename = makeFilename(ext)
	h := make(http.Header, 1)
	h.Set("Content-Type", ct)
	signedURL, err = is.GetSignedURL(ctx, filename, h)
	return
}

func makeFilename(ext string) string {
	var sb strings.Builder
	sb.Grow(len("2006/01/123456789abcdefg.") + len(ext))
	sb.WriteString(time.Now().Format("2006/01/"))
	sb.Write(crockford.Time(crockford.Lower, time.Now()))
	sb.Write(crockford.AppendRandom(crockford.Lower, nil))
	sb.WriteString(".")
	sb.WriteString(ext)
	return sb.String()
}

func hashURLpath(srcPath, ext string) string {
	return fmt.Sprintf("external/%s.%s",
		crockford.AppendMD5(crockford.Lower, nil, []byte(srcPath)),
		ext,
	)
}

func UploadFromURL(ctx context.Context, c *http.Client, is aws.BlobStore, srcurl string) (filename, ext string, err error) {
	res, err := ctxhttp.Get(ctx, c, srcurl)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	const (
		megabyte = 1 << 20
		maxSize  = 25 * megabyte
		peekSize = 512
	)
	buf := bufio.NewReader(http.MaxBytesReader(nil, res.Body, maxSize))

	peek, err := buf.Peek(peekSize)
	if err != nil && err != io.EOF {
		return "", "", err
	}

	ct := mimetype.Detect(peek)
	if ct.Is("image/jpeg") {
		ext = "jpeg"
	} else if ct.Is("image/png") {
		ext = "png"
	} else {
		return "", "", resperr.WithCodeAndMessage(
			fmt.Errorf("%q did not have proper MIME type", srcurl),
			http.StatusBadRequest,
			"URL must be an image",
		)
	}

	slurp, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", "", err
	}

	filename = hashURLpath(srcurl, ext)

	h := make(http.Header, 1)
	h.Set("Content-Type", ct.String())
	if err = is.WriteFile(ctx, filename, h, slurp); err != nil {
		return "", "", resperr.WithCodeAndMessage(
			fmt.Errorf("unexpected S3 status: %d", res.StatusCode),
			http.StatusBadGateway,
			"Could not upload image from URL",
		)
	}
	return filename, ext, nil
}
