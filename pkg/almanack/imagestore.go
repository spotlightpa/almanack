package almanack

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/earthboundkid/crockford/v2"
	"github.com/spotlightpa/almanack/internal/aws"
)

func GetSignedImageUpload(ctx context.Context, is aws.BlobStore, ct string) (signedURL, filename string, err error) {
	filename = makeImageName(ct)
	h := make(http.Header, 1)
	h.Set("Content-Type", ct)
	signedURL, err = is.SignPutURL(ctx, filename, h)
	return
}

func makeImageName(ct string) string {
	now := time.Now().UTC()
	ext := "bin"
	if _, tempext, ok := strings.Cut(ct, "/"); ok && len(tempext) >= 3 {
		ext = tempext
	}
	var sb strings.Builder
	sb.Grow(len("2006/01/1234-5678-9abc-defg.") + len(ext))
	sb.WriteString(now.Format("2006/01/"))
	buf := make([]byte, 0, len("1234-5678-9abc-defg"))
	buf = crockford.AppendTime(crockford.Lower, now, buf)
	buf = crockford.AppendRandom(crockford.Lower, buf)
	buf = crockford.AppendPartition(buf[:0], buf, 4)
	sb.Write(buf)
	sb.WriteByte('.')
	sb.WriteString(ext)
	return sb.String()
}
