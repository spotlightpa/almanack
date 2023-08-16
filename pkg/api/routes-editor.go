package api

import (
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/paginate"
)

func (app *appEnv) userInfo(w http.ResponseWriter, r *http.Request) http.Handler {
	app.logStart(r)
	userinfo := netlifyid.FromContext(r.Context())
	return app.jsonOK(userinfo)
}

func (app *appEnv) getSignupURL(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	app.replyJSON(http.StatusOK, w, app.svc.MailchimpSignupURL)
}

func (app *appEnv) listSharedArticles(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var page int32
	_ = intFromQuery(r, "page", &page)
	if page < 0 {
		app.replyErr(w, r, resperr.WithUserMessage(nil, "Invalid page"))
		return
	}

	// Spotlight PA users can add ?show=all
	if r.URL.Query().Get("show") == "all" {
		if err := app.auth.HasRole(r, "Spotlight PA"); err != nil {
			app.replyErr(w, r, err)
			return
		}

		pager := paginate.PageNumber(page)
		pager.PageSize = 50
		stories, err := paginate.List(pager, r.Context(),
			app.svc.Queries.ListSharedArticles,
			db.ListSharedArticlesParams{
				Offset: pager.Offset(),
				Limit:  pager.Limit(),
			})
		if err != nil {
			app.replyErr(w, r, err)
			return
		}

		app.replyJSON(http.StatusOK, w, struct {
			Stories  []db.SharedArticle `json:"stories"`
			NextPage int32              `json:"next_page,string,omitempty"`
		}{
			Stories:  stories,
			NextPage: pager.NextPage,
		})
		return
	}

	pager := paginate.PageNumber(page)
	pager.PageSize = 20
	stories, err := paginate.List(pager, r.Context(),
		app.svc.Queries.ListSharedArticlesWhereActive,
		db.ListSharedArticlesWhereActiveParams{
			Offset: pager.Offset(),
			Limit:  pager.Limit(),
		})
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, struct {
		Stories  []db.SharedArticle `json:"stories"`
		NextPage int32              `json:"next_page,string,omitempty"`
	}{
		Stories:  stories,
		NextPage: pager.NextPage,
	})
}

func (app *appEnv) getSharedArticle(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var (
		article db.SharedArticle
		err     error
	)
	q := r.URL.Query()
	if st := q.Get("source_type"); st != "" {
		article, err = app.svc.Queries.GetSharedArticleBySource(r.Context(),
			db.GetSharedArticleBySourceParams{
				SourceType: st,
				SourceID:   q.Get("source_id"),
			})
	} else {
		var id int64
		if !intFromQuery(r, "id", &id) {
			app.replyErr(w, r, resperr.WithUserMessage(nil,
				"Must provide article ID"))
			return
		}
		article, err = app.svc.Queries.GetSharedArticleByID(r.Context(), id)
	}
	if err != nil {
		err = db.NoRowsAs404(err,
			"missing shared_article %v", q)
		app.replyErr(w, r, err)
		return
	}

	if article.Status != "S" &&
		article.Status != "P" {
		// Let Spotlight PA users get article regardless of its status
		if err := app.auth.HasRole(r, "Spotlight PA"); err != nil {
			app.replyNewErr(http.StatusNotFound, w, r,
				"user unauthorized to view article: %w", err)
			return
		}
	}

	val, err := app.svc.InflateSharedArticle(r.Context(), &article)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, val)
}
