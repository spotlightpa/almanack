package api

import (
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func (app *appEnv) userInfo(w http.ResponseWriter, r *http.Request) {
	app.Println("start userInfo")
	userinfo, err := netlifyid.FromRequest(r)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, userinfo)
}

func (app *appEnv) listAvailableArcStories(w http.ResponseWriter, r *http.Request) {
	page, err := app.getRequestPage(r, "listAvailableArcStories")
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.Printf("starting listAvailableArcStories page %d", page)

	var res struct {
		Contents []almanack.ArcStory `json:"contents"`
		NextPage int                 `json:"next_page,string,omitempty"`
	}
	if res.Contents, res.NextPage, err = app.svc.ListAvailableArcStories(
		r.Context(), page,
	); err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, res)
}

func (app *appEnv) getArcStory(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	app.Printf("starting getArcStory %s", articleID)

	article, err := app.svc.GetArcStory(r.Context(), articleID)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	if article.Status != almanack.StatusAvailable &&
		article.Status != almanack.StatusPlanned {
		// Let Spotlight PA users get article regardless of its status
		if err := app.auth.HasRole(r, "Spotlight PA"); err != nil {
			app.replyErr(w, r, resperr.New(
				http.StatusNotFound, "user unauthorized to view article: %w", err,
			))
			return
		}
	}

	app.replyJSON(http.StatusOK, w, article)
}

func (app *appEnv) getSignupURL(w http.ResponseWriter, r *http.Request) {
	app.Println("start getSignupURL")
	app.replyJSON(http.StatusOK, w, app.mailchimpSignupURL)
}
