package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/carlmjohnson/workgroup"
	"github.com/go-chi/chi/v5"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/paginate"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (app *appEnv) backgroundSleep(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	l := almlog.FromContext(r.Context())
	if deadline, ok := r.Context().Deadline(); ok {
		l.InfoCtx(r.Context(), "backgroundSleep", "deadline", deadline)
	} else {
		l.InfoCtx(r.Context(), "backgroundSleep", "deadline", false)
	}
	durationStr := chi.URLParam(r, "duration")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	time.Sleep(duration)
	app.replyJSON(http.StatusOK, w, struct {
		SleptFor time.Duration `json:"slept-for"`
	}{duration})
}

func (app *appEnv) backgroundCron(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	if err := workgroup.DoFuncs(workgroup.MaxProcs,
		func() error {
			var errs []error
			// Publish any scheduled pages before pushing new site config
			poperr, warning := app.svc.PopScheduledPages(r.Context())
			if warning != nil {
				app.logErr(r.Context(), warning)
			}
			errs = append(errs, poperr)
			// TODO: Query all locations from DB side
			errs = append(errs, app.svc.PopScheduledSiteChanges(r.Context(), almanack.ElectionFeatLoc))
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
	); err != nil {
		// reply shows up in dev only
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusAccepted, w, "OK")
}

func (app *appEnv) backgroundRefreshPages(w http.ResponseWriter, r *http.Request) {
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
				app.replyErr(w, r, err)
				return
			}
			err = workgroup.DoTasks(workgroup.MaxProcs, pageIDs, func(id int64) error {
				return app.svc.RefreshPageContents(r.Context(), id)
			})
			if err != nil {
				app.replyErr(w, r, err)
				return
			}
			count += len(pageIDs)
			l.Info("backgroundRefreshPages", "processed", count)
		}
	}

	app.replyJSON(http.StatusAccepted, w, "OK")
}

func (app *appEnv) backgroundImages(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	err := workgroup.DoFuncs(workgroup.MaxProcs,
		func() error {
			return app.svc.UploadPendingImages(r.Context())
		},
		func() error {
			return app.svc.ProcessGDocs(r.Context())
		},
	)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusAccepted, w, "OK")
}
