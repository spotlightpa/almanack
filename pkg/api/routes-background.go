package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/carlmjohnson/flowmatic"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/paginate"
	"github.com/spotlightpa/almanack/internal/timex"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (app *appEnv) backgroundSleep(w http.ResponseWriter, r *http.Request) http.Handler {
	app.logStart(r)
	l := almlog.FromContext(r.Context())
	if deadline, ok := r.Context().Deadline(); ok {
		l.InfoContext(r.Context(), "backgroundSleep", "deadline", deadline)
	} else {
		l.InfoContext(r.Context(), "backgroundSleep", "deadline", false)
	}
	durationStr := r.PathValue("duration")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return app.jsonErr(err)
	}
	ok := timex.Sleep(r.Context(), duration)
	return app.jsonAccepted(struct {
		SleptFor time.Duration `json:"slept-for"`
		OK       bool          `json:"ok"`
	}{duration, ok})
}

func (app *appEnv) backgroundCron(w http.ResponseWriter, r *http.Request) http.Handler {
	app.logStart(r)

	if err := app.svc.HC.Start(r.Context()); err != nil {
		app.logErr(r.Context(), err)
		l := almlog.FromContext(r.Context())
		l.ErrorContext(r.Context(), "could not contact Healthchecks.io with start; continuing")
		// fallthrough and keep trying…
	}
	defer func() {
		if suberr := app.svc.HC.Success(r.Context(), nil); suberr != nil {
			app.logErr(r.Context(), suberr)
			l := almlog.FromContext(r.Context())
			l.ErrorContext(r.Context(), "could not contact Healthchecks.io with success")
		}
	}()

	if err := flowmatic.Do(
		func() error {
			var errs []error
			// Publish any scheduled pages before pushing new site config
			poperr, warning := app.svc.PopScheduledPages(r.Context())
			if warning != nil {
				app.logErr(r.Context(), warning)
			}
			errs = append(errs, poperr)
			// TODO: Query all locations from DB side
			errs = append(errs, app.svc.PopScheduledSiteChanges(r.Context(), almanack.HomepageLoc))
			errs = append(errs, app.svc.PopScheduledSiteChanges(r.Context(), almanack.SidebarLoc))
			errs = append(errs, app.svc.PopScheduledSiteChanges(r.Context(), almanack.SiteParamsLoc))
			errs = append(errs, app.svc.PopScheduledSiteChanges(r.Context(), almanack.StateCollegeLoc))
			return errors.Join(errs...)
		},
		func() error {
			return app.svc.UpdateMostPopular(r.Context())
		},
		func() error {
			types, err := app.svc.Queries.ListNewsletterTypes(r.Context())
			if err != nil {
				return err
			}
			var errs []error
			// Update newsletter archives first and then import anything new
			errs = append(errs, app.svc.UpdateNewsletterArchives(r.Context(), types))
			errs = append(errs, app.svc.ImportNewsletterPages(r.Context(), types))
			return errors.Join(errs...)
		},
		func() error {
			return app.svc.UploadPendingImages(r.Context())
		},
		func() error {
			return app.svc.ProcessGDocs(r.Context())
		},
		func() error {
			return app.svc.Queries.DeleteGDocsDocWhereUnunused(r.Context())
		},
		func() error {
			return updateMD5s(
				r.Context(),
				app.svc.Queries.ListImagesWhereNoMD5,
				func(ctx context.Context, image db.Image) ([]byte, int64, error) {
					return app.svc.ImageStore.ReadMD5(ctx, image.Path)
				},
				func(ctx context.Context, hash []byte, size int64, image db.Image) error {
					_, err := app.svc.Queries.UpdateImageMD5Size(r.Context(), db.UpdateImageMD5SizeParams{
						ID:    image.ID,
						MD5:   hash,
						Bytes: size,
					})
					return err
				},
			)
		},
		func() error {
			return updateMD5s(
				r.Context(),
				app.svc.Queries.ListFilesWhereNoMD5,
				func(ctx context.Context, file db.File) ([]byte, int64, error) {
					path := strings.TrimPrefix(file.URL, "https://files.data.spotlightpa.org/")
					return app.svc.FileStore.ReadMD5(ctx, path)
				},
				func(ctx context.Context, hash []byte, size int64, file db.File) error {
					_, err := app.svc.Queries.UpdateFileMD5Size(r.Context(), db.UpdateFileMD5SizeParams{
						ID:    file.ID,
						MD5:   hash,
						Bytes: size,
					})
					return err
				},
			)
		},
	); err != nil {
		// reply shows up in dev only
		return app.jsonErr(err)
	}

	return app.jsonAccepted("OK")
}

func (app *appEnv) backgroundRefreshPages(w http.ResponseWriter, r *http.Request) http.Handler {
	app.logStart(r)
	l := almlog.FromContext(r.Context())
	count := 0
	for _, filepath := range []string{
		"content/news/%",
		"content/statecollege/%",
	} {
		pager := paginate.PageNumber[int32](0)
		pager.PageSize = 10

		for pager.HasMore() {
			pager.Advance()
			pageIDs, err := paginate.List(
				pager, r.Context(),
				app.svc.Queries.ListPageIDs,
				db.ListPageIDsParams{
					FilePath: filepath,
					Offset:   pager.Offset(),
					Limit:    pager.Limit(),
				})
			if err != nil {
				return app.jsonErr(err)
			}
			err = flowmatic.Each(flowmatic.MaxProcs, pageIDs, func(id int64) error {
				return app.svc.RefreshPageContents(r.Context(), id)
			})
			if err != nil {
				return app.jsonErr(err)
			}
			count += len(pageIDs)
			l.Info("backgroundRefreshPages", "processed", count)
		}
	}

	return app.jsonAccepted("OK")
}

func (app *appEnv) backgroundImages(w http.ResponseWriter, r *http.Request) http.Handler {
	app.logStart(r)

	err := flowmatic.Do(
		func() error {
			return app.svc.UploadPendingImages(r.Context())
		},
		func() error {
			return app.svc.ProcessGDocs(r.Context())
		},
	)
	if err != nil {
		return app.jsonErr(err)
	}

	return app.jsonAccepted("OK")
}

func updateMD5s[T any](
	ctx context.Context,
	list func(context.Context, int32) ([]T, error),
	read func(context.Context, T) ([]byte, int64, error),
	update func(context.Context, []byte, int64, T) error,
) error {
	for {
		items, err := list(ctx, 10)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		if err = flowmatic.Each(5, items, func(item T) error {
			hash, size, err := read(ctx, item)
			if err != nil {
				return err
			}
			return update(ctx, hash, size, item)
		}); err != nil {
			return err
		}
	}
}
