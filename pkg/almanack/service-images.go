package almanack

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/workgroup"
	"github.com/spotlightpa/almanack/internal/db"
)

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
