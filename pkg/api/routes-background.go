package api

import (
	"net/http"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/go-chi/chi"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func (app *appEnv) backgroundSleep(w http.ResponseWriter, r *http.Request) {
	app.Println("start backgroundSleep")
	if deadline, ok := r.Context().Deadline(); ok {
		app.Printf("deadline: %s", deadline.Format(time.RFC1123))
	} else {
		app.Printf("no deadline")
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
	app.Println("start background cron")

	if err := errutil.ExecParallel(func() error {
		var errs errutil.Slice
		// Publish any scheduled pages before pushing new site config
		poperr, warning := app.svc.PopScheduledPages(r.Context())
		app.logErr(r.Context(), warning)
		errs.Push(poperr)
		errs.Push(app.svc.PopScheduledSiteChanges(r.Context(), almanack.EditorsPicksLoc))
		errs.Push(app.svc.PopScheduledSiteChanges(r.Context(), almanack.SiteParamsLoc))
		return errs.Merge()
	}, func() error {
		return app.svc.UpdateMostPopular(r.Context())
	}, func() error {
		var errs errutil.Slice
		// Update newsletter archives first and then import anything new
		errs.Push(app.svc.UpdateNewsletterArchives(r.Context()))
		errs.Push(app.svc.ImportNewsletterPages(r.Context()))
		return errs.Merge()
	}); err != nil {
		// reply shows up in dev only
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusAccepted, w, "OK")
}

func (app *appEnv) backgroundRefreshPages(w http.ResponseWriter, r *http.Request) {
	app.Println("start backgroundRefreshPages")

	hasMore := true
	for queryPage := int32(0); hasMore; queryPage++ {
		pager := db.PageNumSize(queryPage, 10)
		pageIDs, err := db.Paginate(
			pager, r.Context(),
			app.svc.Queries.ListPageIDs,
			db.ListPageIDsParams{
				FilePath: "content/news/%",
				Offset:   pager.Offset(),
				Limit:    pager.Limit(),
			})
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		for _, id := range pageIDs {
			if err := app.svc.RefreshPageContents(r.Context(), id); err != nil {
				app.replyErr(w, r, err)
				return
			}
		}
		hasMore = pager.HasMore()
	}

	app.replyJSON(http.StatusAccepted, w, "OK")
}
