package almanack

import (
	"fmt"
	"strings"
	"time"

	"github.com/carlmjohnson/crockford"
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
