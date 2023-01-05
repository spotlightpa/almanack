package almanack

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/carlmjohnson/errutil"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/common"
)

func (svc Services) ReplaceImageURL(ctx context.Context, srcURL, description, credit string) (s string, err error) {
	defer errutil.Trace(&err)

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

	ext := path.Ext(srcURL)
	ext = strings.TrimPrefix(ext, ".")
	uploadPath := hashURLpath(srcURL, ext)
	itype, err := svc.Queries.GetImageTypeForExtension(ctx, ext)
	if err != nil {
		if db.IsNotFound(err) {
			return "", fmt.Errorf("unknown image extension (%q) on source: %q",
				ext, srcURL)
		}
		return "", err
	}
	_, err = svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
		Path:        uploadPath,
		Type:        itype.Name,
		Description: description,
		Credit:      credit,
		SourceURL:   srcURL,
		IsUploaded:  false,
	})
	return uploadPath, err
}

func (svc Services) UpdateMostPopular(ctx context.Context) (err error) {
	defer errutil.Trace(&err)

	common.Logger.Printf("updating most popular")
	cl, err := svc.gsvc.GAClient(ctx)
	if err != nil {
		return err
	}
	pages, err := svc.gsvc.MostPopularNews(ctx, cl)
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
