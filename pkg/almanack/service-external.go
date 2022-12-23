package almanack

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/common"
)

func (svc Services) ReplaceImageURL(ctx context.Context, srcURL, description, credit string) (s string, err error) {
	defer errutil.Trace(&err)

	if srcURL == "" {
		return "", fmt.Errorf("no image provided")
	}
	image, err := svc.Queries.GetImageBySourceURL(ctx, srcURL)
	if err != nil && !db.IsNotFound(err) {
		return "", err
	}
	if !db.IsNotFound(err) && image.IsUploaded {
		return image.Path, nil
	}
	var path, ext string
	if path, ext, err = UploadFromURL(ctx, svc.Client, svc.ImageStore, srcURL); err != nil {
		return "", resperr.New(
			http.StatusBadGateway,
			"could not upload image %s: %w", srcURL, err,
		)
	}
	_, err = svc.Queries.UpsertImage(ctx, db.UpsertImageParams{
		Path:        path,
		Type:        ext,
		Description: description,
		Credit:      credit,
		SourceURL:   srcURL,
		IsUploaded:  true,
	})
	return path, err
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
