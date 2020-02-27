package almanack

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/crockford"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

type ImageStore interface {
	GetSignedURL(srcPath string) (signedURL string, err error)
}

func GetSignedUpload(is ImageStore) (signedURL, filename string, err error) {
	filename = makeFilename()
	signedURL, err = is.GetSignedURL(filename)
	return
}

func GetSignedHashedUrl(is ImageStore, srcurl string) (signedURL, filename string, err error) {
	filename = hashURLpath(srcurl)
	signedURL, err = is.GetSignedURL(filename)
	return
}

func makeFilename() string {
	var sb strings.Builder
	sb.Grow(len("2006/01/123456789abcdefg.jpeg"))
	sb.WriteString(time.Now().Format("2006/01/"))
	sb.Write(crockford.Time(crockford.Lower, time.Now()))
	sb.Write(crockford.AppendRandom(crockford.Lower, nil))
	sb.WriteString(".jpeg")
	return sb.String()
}

func hashURLpath(srcPath string) string {
	return fmt.Sprintf("external/%s.jpeg",
		crockford.AppendMD5(crockford.Lower, nil, []byte(srcPath)))
}

func UploadFromURL(ctx context.Context, c *http.Client, is ImageStore, srcurl string) (filename string, err error) {
	if c == nil {
		c = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(
		ctx, "GET", srcurl, nil)
	if err != nil {
		return "", err
	}
	res, err := c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	const (
		megabyte = 1 << 20
		maxSize  = 25 * megabyte
	)
	buf := bufio.NewReader(http.MaxBytesReader(nil, res.Body, maxSize))
	// http.DetectContentType only uses first 512 bytes
	peek, err := buf.Peek(512)
	if err != nil && err != io.EOF {
		return "", err
	}

	if ct := http.DetectContentType(peek); !strings.HasPrefix(ct, "image/jpeg") {
		return "", errutil.Response{
			StatusCode: http.StatusBadRequest,
			Message:    "URL must be an image",
			Log: fmt.Sprintf("%q did not have proper MIME type",
				srcurl),
		}
	}
	slurp, err := ioutil.ReadAll(buf)
	if err != nil {
		return "", err
	}

	var signedURL string
	signedURL, filename, err = GetSignedHashedUrl(is, srcurl)
	if err != nil {
		return "", err
	}

	req, err = http.NewRequestWithContext(
		ctx, "PUT", signedURL, bytes.NewReader(slurp))
	if err != nil {
		return "", err
	}
	res, err = c.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errutil.Response{
			StatusCode: http.StatusBadGateway,
			Message:    "Could not upload image from URL",
			Log:        fmt.Sprintf("unexpected S3 status: %d", res.StatusCode),
		}
	}
	return filename, nil
}
