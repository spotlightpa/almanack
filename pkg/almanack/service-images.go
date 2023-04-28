package almanack

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/carlmjohnson/crockford"
	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/carlmjohnson/workgroup"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/google"
)

func FetchImageURL(ctx context.Context, c *http.Client, srcurl string) (body []byte, ctype string, err error) {
	var buf bytes.Buffer
	if err = requests.
		URL(srcurl).
		Client(c).
		CheckStatus(http.StatusOK).
		CheckPeek(512, func(peek []byte) error {
			ct := mimetype.Detect(peek)
			if ct.Is("image/jpeg") ||
				ct.Is("image/png") ||
				ct.Is("image/tiff") ||
				ct.Is("image/webp") ||
				ct.Is("image/avif") ||
				ct.Is("image/heic") {
				ctype = ct.String()
				return nil
			}
			return resperr.WithUserMessage(
				fmt.Errorf("%q did not have proper MIME type: %s",
					srcurl, ct.String()),
				"URL must be an image",
			)
		}).
		ToBytesBuffer(&buf).
		Fetch(ctx); err != nil {
		if resperr.StatusCode(err) != http.StatusInternalServerError {
			return nil, "", err
		}
		return nil, "", resperr.New(http.StatusBadGateway, "problem fetching image: %w", err)
	}

	return buf.Bytes(), ctype, nil
}

func (svc Services) ReplaceImageURL(ctx context.Context, srcURL, description, credit string) (s string, err error) {
	defer errorx.Trace(&err)

	if srcURL == "" {
		return "", fmt.Errorf("no image provided")
	}
	image, err := svc.Queries.GetImageBySourceURL(ctx, srcURL)
	if err == nil { // found entry
		return image.Path, nil
	}
	if !db.IsNotFound(err) {
		return "", err
	}

	itype, err := svc.imageTypeFromExtension(ctx, srcURL)
	if err != nil {
		return "", err
	}
	uploadPath := hashURLpath(srcURL, itype)
	_, err = svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
		Path:        uploadPath,
		Type:        itype,
		Description: description,
		Credit:      credit,
		SourceURL:   srcURL,
		IsUploaded:  false,
	})
	return uploadPath, err
}

func (svc Services) ReplaceAndUploadImageURL(ctx context.Context, srcURL, description, credit string) (path string, err error) {
	defer errorx.Trace(&err)

	if srcURL == "" {
		return "", fmt.Errorf("no image provided")
	}
	image, err := svc.Queries.GetImageBySourceURL(ctx, srcURL)
	switch {
	case err == nil: // found entry
		return image.Path, nil
	case !db.IsNotFound(err): // unexpected DB problem
		return "", err
	}

	cl := svc.Client
	dlURL := srcURL
	// See if it's a Drive URL
	if id, err := google.NormalizeFileID(srcURL); err == nil {
		if err = svc.ConfigureGoogleCert(ctx); err != nil {
			return "", err
		}
		cl, err = svc.Gsvc.DriveClient(ctx)
		if err != nil {
			return "", err
		}
		if dlURL, err = svc.Gsvc.DownloadURLForDriveID(id); err != nil {
			return "", err
		}
	}
	// Download the image + headers
	body, ct, err := FetchImageURL(ctx, cl, dlURL)
	if err != nil {
		return "", err
	}

	itype, err := imageTypeFromMIME(ct)
	if err != nil {
		return "", err
	}

	uploadPath := hashURLpath(srcURL, itype)
	if _, err = svc.uploadAndRecordImage(ctx, uploadAndRecordImageParams{
		UploadPath:  uploadPath,
		Body:        body,
		ContentType: ct,
		Description: description,
		Credit:      credit,
		SourceURL:   srcURL,
	}); err != nil {
		return "", err
	}

	return uploadPath, nil
}

func hashURLpath(srcPath, ext string) string {
	return fmt.Sprintf("external/%s.%s",
		crockford.MD5(crockford.Lower, []byte(srcPath)),
		ext,
	)
}

type uploadAndRecordImageParams struct {
	UploadPath  string
	Body        []byte
	ContentType string
	Description string
	Credit      string
	SourceURL   string
}

func (svc Services) uploadAndRecordImage(ctx context.Context, arg uploadAndRecordImageParams) (*db.Image, error) {
	itype, err := imageTypeFromMIME(arg.ContentType)
	if err != nil {
		return nil, err
	}

	h := http.Header{"Content-Type": []string{arg.ContentType}}
	if err := svc.ImageStore.WriteFile(ctx, arg.UploadPath, h, arg.Body); err != nil {
		return nil, err
	}

	dbImage, err := svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
		Path:        arg.UploadPath,
		Type:        itype,
		Description: arg.Description,
		Credit:      arg.Credit,
		SourceURL:   arg.SourceURL,
		IsUploaded:  true,
	})
	if err != nil {
		return nil, err
	}
	return &dbImage, nil
}

func (svc Services) UploadPendingImages(ctx context.Context) error {
	images, err := svc.Queries.ListImageWhereNotUploaded(ctx)
	if err != nil {
		return err
	}

	return workgroup.DoTasks(5, images,
		func(image db.Image) error {
			return svc.uploadPendingImage(ctx, image.SourceURL, image.Path)
		})
}

func (svc Services) uploadPendingImage(ctx context.Context, sourceURL, path string) (err error) {
	defer errorx.Trace(&err)

	body, ctype, err := FetchImageURL(ctx, svc.Client, sourceURL)
	if err != nil {
		return err
	}

	h := http.Header{"Content-Type": []string{ctype}}
	if err = svc.ImageStore.WriteFile(ctx, path, h, body); err != nil {
		return fmt.Errorf("uploadPendingImage: ImageStore.WriteFile: %w", err)
	}
	if err != nil {
		return err
	}
	if _, err = svc.Queries.UpdateImage(ctx,
		db.UpdateImageParams{
			Path:      path,
			SourceURL: sourceURL,
		}); err != nil {
		return err
	}
	return nil
}

func imageTypeFromMIME(ct string) (string, error) {
	itype, ok := strings.CutPrefix(ct, "image/")
	if !ok {
		return "", fmt.Errorf("bad image Content-Type: %q", ct)
	}
	return itype, nil
}

func (svc Services) imageTypeFromExtension(ctx context.Context, srcURL string) (typeName string, err error) {
	ext := path.Ext(srcURL)
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)
	itype, err := svc.Queries.GetImageTypeForExtension(ctx, ext)
	if err != nil {
		if db.IsNotFound(err) {
			return "", resperr.WithUserMessagef(err,
				"Unknown image extension (%q) on source: %q",
				ext, srcURL)
		}
		return "", err
	}
	return itype.Name, nil
}
