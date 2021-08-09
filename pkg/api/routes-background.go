package api

import (
	"net/http"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/go-chi/chi"
	"github.com/spotlightpa/almanack/internal/db"
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
		return app.svc.PopScheduledArticles(r.Context())
	}, func() error {
		return app.svc.PopScheduledPages(r.Context())
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

	hasNext := true
	for queryPage := int32(0); hasNext; queryPage++ {
		const (
			pageSize = 10
			limit    = pageSize + 1
		)
		offset := queryPage * pageSize
		pages, err := app.svc.Queries.ListPages(r.Context(), db.ListPagesParams{
			FilePath: "content/news/%",
			Offset:   offset,
			Limit:    limit,
		})
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		hasNext = len(pages) == limit
		if hasNext {
			pages = pages[:pageSize]
		}
		for i := range pages {
			if err := app.svc.RefreshPageContents(r.Context(), pages[i].ID); err != nil {
				app.replyErr(w, r, err)
				return
			}
		}
	}

	app.replyJSON(http.StatusAccepted, w, "OK")
}
