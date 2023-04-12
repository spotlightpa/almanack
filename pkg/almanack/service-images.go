package almanack

import (
	"context"

	"github.com/carlmjohnson/workgroup"
	"github.com/spotlightpa/almanack/internal/db"
)

func (svc Services) UploadImages(ctx context.Context) error {
	images, err := svc.Queries.ListImageWhereNotUploaded(ctx)
	if err != nil {
		return err
	}

	return workgroup.DoTasks(5, images,
		func(image db.Image) error {
			_, _, err := UploadFromURL(
				ctx,
				svc.Client,
				svc.ImageStore,
				image.Path,
				image.SourceURL)
			if err != nil {
				return err
			}
			if _, err = svc.Queries.UpdateImage(ctx,
				db.UpdateImageParams{
					Path:      image.Path,
					SourceURL: image.SourceURL,
				}); err != nil {
				return err
			}
			return nil
		})
}
