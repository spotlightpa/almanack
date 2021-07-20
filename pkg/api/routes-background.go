package api

import (
	"net/http"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/go-chi/chi"
)

func (app *appEnv) backgroundSleep(w http.ResponseWriter, r *http.Request) {
	app.Println("start backgroundSleep")
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

	var errs errutil.Slice
	// Update newsletter archives first and then import anything new
	errs.Push(app.svc.UpdateNewsletterArchives(r.Context()))
	errs.Push(app.svc.ImportNewsletterPages(r.Context()))

	if err := errs.Merge(); err != nil {
		// reply shows up in dev only
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusAccepted, w, "OK")
}
