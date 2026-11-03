package almsvc

import (
	"context"
	"crypto/md5"
	"net/http"
	"strings"

	"github.com/earthboundkid/crockford/v2"
	"github.com/earthboundkid/errorx/v2"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/utils/pgxutil"
)

type UploadGDocsImageParams struct {
	ExternalID  string
	DocObjectID string
	ImageURL    string
	Embed       *db.EmbedImage // In-out param
}

func (svc Services) UploadGDocsImage(ctx context.Context, arg UploadGDocsImageParams) (err error) {
	defer errorx.Trace(&err)

	// Download the image + headers
	body, ct, err := FetchImageURL(ctx, svc.Client, arg.ImageURL)
	if err != nil {
		return err
	}

	// Hash the file
	hash := md5.Sum(body)

	// Look up file hash
	dbImage, err := svc.Queries.GetImageByMD5(ctx, hash[:])
	var imageID int64
	switch {
	// If it is found, return & save the relationship for next refresh
	case err == nil:
		arg.Embed.Path = dbImage.Path
		imageID = dbImage.ID

	// If it's not found, it needs to be uploaded & saved
	case pgxutil.IsNotFound(err):
		itype, err := imageTypeFromMIME(ct)
		if err != nil {
			return err
		}
		uploadPath := makeCASaddress(body, ct)
		arg.Embed.Path = uploadPath
		h := http.Header{"Content-Type": []string{ct}}
		if err := svc.ImageStore.WriteFile(ctx, uploadPath, h, body); err != nil {
			return err
		}
		record, err := svc.Queries.UpsertImageWithMD5(ctx, db.UpsertImageWithMD5Params{
			Path:        uploadPath,
			Type:        itype,
			Description: arg.Embed.Description,
			Credit:      arg.Embed.Credit,
			MD5:         hash[:],
			Bytes:       int64(len(hash)),
		})
		if err != nil {
			return err
		}
		imageID = record.ID
	// Other errors are bad
	case err != nil:
		return err
	}

	return svc.Queries.UpsertGDocsImage(ctx, db.UpsertGDocsImageParams{
		ExternalID:  arg.ExternalID,
		DocObjectID: arg.DocObjectID,
		ImageID:     imageID,
	})
}

func makeCASaddress(body []byte, ct string) string {
	// https://en.wikipedia.org/wiki/Content-addressable_storage
	b := make([]byte, 0, crockford.LenMD5)
	b = crockford.AppendMD5(crockford.Lower, b, body)[:16]
	b = crockford.AppendPartition(b[:0], b, 4)
	ext, ok := strings.CutPrefix(ct, "image/")
	if !ok {
		ext = "bin"
	}
	return "cas/" + string(b) + "." + ext
}
