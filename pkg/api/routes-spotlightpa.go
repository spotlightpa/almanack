package api

import (
	"fmt"
	"net/http"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func (app *appEnv) listAllArcStories(w http.ResponseWriter, r *http.Request) {
	page, err := app.getPage(r, "listAllArcStories")
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.Printf("start listAllArcStories page %d", page)

	var resp struct {
		Contents []almanack.ArcStory `json:"contents"`
		NextPage int                 `json:"next_page,omitempty"`
	}

	resp.Contents, resp.NextPage, err = app.svc.ListAllArcStories(r.Context(), page)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, &resp)
}

func (app *appEnv) listWithArcRefresh(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listWithArcRefresh")
	type response struct {
		Contents []almanack.ArcStory `json:"contents"`
		NextPage int                 `json:"next_page,omitempty"`
	}
	var (
		feed *almanack.ArcAPI
		err  error
	)
	if feed, err = app.FetchFeed(r.Context()); err != nil {
		// Keep trucking even if you can't load feed
		app.logErr(r.Context(), err)
	} else if err = app.svc.StoreFeed(r.Context(), feed); err != nil {
		app.replyErr(w, r, err)
		return
	}
	var resp response
	resp.Contents, resp.NextPage, err = app.svc.ListAllArcStories(r.Context(), 0)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, resp)
}

func (app *appEnv) postAlmanackArcStory(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting postAlmanackArcStory")

	var userData struct {
		ID         string          `json:"_id"`
		Note       string          `json:"almanack-note,omitempty"`
		Status     almanack.Status `json:"almanack-status,omitempty"`
		RefreshArc bool            `json:"almanack-refresh-arc"`
	}
	if !app.readJSON(w, r, &userData) {
		return
	}

	var (
		story        almanack.ArcStory
		refreshStory bool
	)
	if userData.RefreshArc {
		var (
			feed *almanack.ArcAPI
			err  error
		)
		if feed, err = app.FetchFeed(r.Context()); err != nil {
			app.replyErr(w, r, err)
			return
		}
		if err := app.svc.StoreFeed(r.Context(), feed); err != nil {
			app.replyErr(w, r, err)
			return
		}
		for i := range feed.Contents {
			if feed.Contents[i].ID == userData.ID {
				story.ArcFeedItem = feed.Contents[i]
				refreshStory = true
			}
		}
	}
	story.ID = userData.ID
	story.Note = userData.Note
	story.Status = userData.Status

	if err := app.svc.SaveAlmanackArticle(r.Context(), &story, refreshStory); err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusAccepted, w, &userData)
}

func (app *appEnv) postMessage(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting postMessage")
	type request struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	var req request
	if !app.readJSON(w, r, &req) {
		return
	}
	if err := app.email.SendEmail(req.Subject, req.Body); err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusAccepted, w, http.StatusText(http.StatusAccepted))
}

func (app *appEnv) getScheduledArticle(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	app.Printf("start getScheduledArticle %s", articleID)

	article, err := app.svc.GetScheduledArticle(r.Context(), articleID)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, article)
}

func (app *appEnv) postScheduledArticle(w http.ResponseWriter, r *http.Request) {
	app.Println("start postScheduledArticle")

	var userData struct {
		almanack.SpotlightPAArticle
		RefreshArc bool `json:"almanack-refresh-arc"`
	}
	if !app.readJSON(w, r, &userData) {
		return
	}

	if userData.RefreshArc {
		var (
			feed *almanack.ArcAPI
			err  error
		)
		if feed, err = app.FetchFeed(r.Context()); err != nil {
			app.replyErr(w, r, err)
			return
		}
		if err = app.svc.StoreFeed(r.Context(), feed); err != nil {
			app.replyErr(w, r, err)
			return
		}
		if err = app.svc.ResetSpotlightPAArticleArcData(r.Context(), &userData.SpotlightPAArticle); err != nil {
			app.replyErr(w, r, err)
			return
		}
	}

	if err := app.svc.SaveScheduledArticle(r.Context(), &userData.SpotlightPAArticle); err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusAccepted, w, &userData.SpotlightPAArticle)
}

var supportedContentTypes = map[string]string{
	"image/jpeg": "jpeg",
	"image/png":  "png",
	"image/tiff": "tiff",
}

func (app *appEnv) postSignedUpload(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postSignedUpload")
	var userData struct {
		Type string `json:"type"`
	}
	if !app.readJSON(w, r, &userData) {
		return
	}

	ext, ok := supportedContentTypes[userData.Type]
	if !ok {
		app.replyErr(w, r, resperr.New(
			http.StatusBadRequest, "unsupported content type: %q", ext,
		))
		return
	}

	type response struct {
		SignedURL string `json:"signed-url"`
		FileName  string `json:"filename"`
	}
	var (
		res response
		err error
	)
	res.SignedURL, res.FileName, err = almanack.GetSignedImageUpload(
		r.Context(), app.svc.ImageStore, userData.Type)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	if n, err := app.svc.Queries.CreateImagePlaceholder(r.Context(), db.CreateImagePlaceholderParams{
		Path: res.FileName,
		Type: ext,
	}); err != nil {
		app.replyErr(w, r, err)
		return
	} else if n != 1 {
		// Log and continue
		app.logErr(r.Context(),
			fmt.Errorf("creating image %q but it already exists", res.FileName))
	}
	app.replyJSON(http.StatusOK, w, &res)
}

func (app *appEnv) postImageUpdate(w http.ResponseWriter, r *http.Request) {
	app.Println("start postImageUpdate")

	var userData db.UpdateImageParams
	if !app.readJSON(w, r, &userData) {
		return
	}
	var (
		res db.Image
		err error
	)
	if res, err = app.svc.Queries.UpdateImage(r.Context(), userData); err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, &res)
}

func (app *appEnv) listDomains(w http.ResponseWriter, r *http.Request) {
	app.Println("start listDomains")
	type response struct {
		Domains []string `json:"domains"`
	}

	domains, err := app.svc.Queries.ListDomainsWithRole(r.Context(), "editor")
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, response{
		domains,
	})
}

func (app *appEnv) postDomain(w http.ResponseWriter, r *http.Request) {
	app.Println("start postDomain")
	type request struct {
		Domain string `json:"domain"`
		Remove bool   `json:"remove"`
	}
	type response struct {
		Domains []string `json:"domains"`
	}
	var req request
	if !app.readJSON(w, r, &req) {
		return
	}

	if req.Domain == "" {
		app.replyErr(w, r, resperr.New(http.StatusBadRequest,
			"can't add nothing!"))
		return
	}

	if req.Domain == "spotlightpa.org" {
		app.replyErr(w, r, resperr.New(http.StatusBadRequest,
			"can't change spotlightpa.org!"))
		return
	}

	var roles []string
	if !req.Remove {
		roles = []string{"editor"}
	}

	if _, err := app.svc.Queries.SetRolesForDomain(
		r.Context(),
		db.SetRolesForDomainParams{
			Domain: req.Domain,
			Roles:  roles,
		},
	); err != nil {
		app.replyErr(w, r, err)
		return
	}

	domains, err := app.svc.Queries.ListDomainsWithRole(r.Context(), "editor")
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, response{
		domains,
	})
}

func (app *appEnv) listAddresses(w http.ResponseWriter, r *http.Request) {
	app.Println("start listAddresses")
	var (
		resp struct {
			Addresses []string `json:"addresses"`
		}
		err error
	)
	resp.Addresses, err = app.svc.Queries.ListAddressesWithRole(r.Context(), "editor")
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, resp)
}

func (app *appEnv) postAddress(w http.ResponseWriter, r *http.Request) {
	app.Println("start postAddresses")
	type request struct {
		Address string `json:"address"`
		Remove  bool   `json:"remove"`
	}
	type response struct {
		Addresses []string `json:"addresses"`
	}
	var req request
	if !app.readJSON(w, r, &req) {
		return
	}

	if !emailx.Valid(req.Address) {
		app.replyErr(w, r, resperr.New(http.StatusBadRequest,
			"invalid email address: %q", req.Address))
		return
	}

	var roles []string
	if !req.Remove {
		roles = []string{"editor"}
	}

	if _, err := app.svc.Queries.SetRolesForAddress(
		r.Context(),
		db.SetRolesForAddressParams{
			EmailAddress: req.Address,
			Roles:        roles,
		},
	); err != nil {
		app.replyErr(w, r, err)
		return
	}

	var (
		resp response
		err  error
	)
	resp.Addresses, err = app.svc.Queries.ListAddressesWithRole(r.Context(), "editor")
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, resp)
}

func (app *appEnv) listSpotlightPAArticles(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listSpotlightPAArticles")
	type response struct {
		Articles []db.ListSpotlightPAArticlesRow `json:"articles"`
	}
	var (
		res response
		err error
	)

	if res.Articles, err = app.svc.Queries.ListSpotlightPAArticles(r.Context()); err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, res)
}

func (app *appEnv) listImages(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listImages")
	type response struct {
		Images []db.Image `json:"images"`
	}
	var (
		res response
		err error
	)

	if res.Images, err = app.svc.Queries.ListImages(r.Context(), db.ListImagesParams{
		Offset: 0,
		Limit:  100,
	}); err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, res)
}

func (app *appEnv) getEditorsPicks(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting getEditorsPicks")
	resp, err := almanack.GetEditorsPicks(r.Context(), app.svc.Queries)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, resp)
}

func (app *appEnv) postEditorsPicks(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting postEditorsPicks")
	var req almanack.EditorsPicks
	if !app.readJSON(w, r, &req) {
		return
	}
	if err := almanack.SetEditorsPicks(
		r.Context(),
		app.svc.Queries,
		app.svc.ContentStore,
		&req,
	); err != nil {
		app.replyErr(w, r, err)
		return
	}
	resp, err := almanack.GetEditorsPicks(r.Context(), app.svc.Queries)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, resp)
}

func (app *appEnv) listAllTopics(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listAllTopics")
	t, err := app.svc.Queries.ListAllTopics(r.Context())
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, struct {
		Topics []string `json:"topics"`
	}{t})
}

func (app *appEnv) listAllSeries(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listAllSeries")
	s, err := app.svc.Queries.ListAllSeries(r.Context())
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, struct {
		Series []string `json:"series"`
	}{s})
}

func (app *appEnv) listFiles(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listFiles")
	type response struct {
		Files []db.File `json:"files"`
	}
	var (
		res response
		err error
	)

	if res.Files, err = app.svc.Queries.ListFiles(r.Context(), db.ListFilesParams{
		Offset: 0,
		Limit:  100,
	}); err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, res)
}

func (app *appEnv) postFileCreate(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postFileCreate")
	var userData struct {
		MimeType string `json:"mimeType"`
		FileName string `json:"filename"`
	}
	if !app.readJSON(w, r, &userData) {
		return
	}
	type response struct {
		SignedURL    string `json:"signed-url"`
		FileURL      string `json:"file-url"`
		Disposition  string `json:"disposition"`
		CacheControl string `json:"cache-control"`
	}
	var (
		res response
		err error
	)

	res.SignedURL, res.FileURL, res.Disposition, res.CacheControl, err = almanack.GetSignedFileUpload(
		r.Context(),
		app.svc.FileStore,
		userData.FileName,
		userData.MimeType,
	)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	if n, err := app.svc.Queries.CreateFilePlaceholder(r.Context(),
		db.CreateFilePlaceholderParams{
			Filename: userData.FileName,
			Type:     userData.MimeType,
			URL:      res.FileURL,
		}); err != nil {
		app.replyErr(w, r, err)
		return
	} else if n != 1 {
		// Log and continue
		app.logErr(r.Context(),
			fmt.Errorf("creating file %q but it already exists", res.FileURL))
	}
	app.replyJSON(http.StatusOK, w, &res)
}

func (app *appEnv) postFileUpdate(w http.ResponseWriter, r *http.Request) {
	app.Println("start postFileUpdate")

	var userData db.UpdateFileParams
	if !app.readJSON(w, r, &userData) {
		return
	}
	var (
		res db.File
		err error
	)
	if res, err = app.svc.Queries.UpdateFile(r.Context(), userData); err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, &res)
}

func (app *appEnv) listNewsletterPages(w http.ResponseWriter, r *http.Request) {
	page, err := app.getPage(r, "listNewsletterPages")
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.Printf("start listNewsletterPages page %d", page)

	var resp struct {
		Pages    []db.ListPagesRow `json:"pages"`
		NextPage int               `json:"next_page,omitempty"`
	}
	const limit = 100
	offset := int32(page) * limit

	resp.Pages, err = app.svc.Queries.ListPages(r.Context(), db.ListPagesParams{
		FilePath: "content/newsletters/%",
		Limit:    limit + 1,
		Offset:   offset,
	})
	hasNext := len(resp.Pages) == limit+1
	if hasNext {
		resp.Pages = resp.Pages[:limit]
		resp.NextPage = page + 1
	}
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, &resp)
}
