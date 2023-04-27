package almanack

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

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

	itype, err := svc.typeForImage(ctx, srcURL)
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

	itype, ok := strings.CutPrefix(ct, "image/")
	if !ok {
		return "", fmt.Errorf("bad image content-type for %s: %q", srcURL, ct)
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

func (svc Services) typeForImage(ctx context.Context, srcURL string) (typeName string, err error) {
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

func (svc Services) UpdateMostPopular(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	l := almlog.FromContext(ctx)
	l.InfoCtx(ctx, "Services.UpdateMostPopular")

	if err = svc.ConfigureGoogleCert(ctx); err != nil {
		return err
	}
	cl, err := svc.Gsvc.GAClient(ctx)
	if err != nil {
		return err
	}
	pages, err := svc.Gsvc.MostPopularNews(ctx, cl)
	if err != nil {
		return err
	}
	if len(pages) < 5 {
		return fmt.Errorf("expected more popular pages; got %q", pages)
	}
	data, err := svc.Queries.ListPagesByURLPaths(ctx, pages)
	if err != nil {
		return err
	}
	if len(data) < 5 {
		return fmt.Errorf("could not find popular pages; got %q from %q",
			data, pages)
	}
	return UploadJSON(
		ctx,
		svc.FileStore,
		"feeds/most-popular-items.json",
		"public, max-age=300",
		struct {
			Pages []db.ListPagesByURLPathsRow `json:"pages"`
		}{
			data,
		},
	)
}

func (svc Services) ConfigureGoogleCert(ctx context.Context) (err error) {
	svc.gsvcOnce.Do(func() {
		defer errorx.Trace(&err)

		if svc.Gsvc.HasCert() {
			return
		}

		opt, err := svc.Queries.GetOption(ctx, "google-json")
		switch {
		case db.IsNotFound(err):
			l := almlog.FromContext(ctx)
			l.Warn("ConfigureGoogleCert: no certificate in database")
			return
		case err != nil:
			svc.gsvcErr = err
			return
		case err == nil:
			svc.gsvcErr = svc.Gsvc.ConfigureCert(opt)
		}
	})
	return svc.gsvcErr
}
