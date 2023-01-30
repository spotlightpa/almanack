package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/resperr"
	"github.com/jackc/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/paginate"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"golang.org/x/exp/slices"
)

func (app *appEnv) postMessage(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	type request struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	var req request
	if !app.readJSON(w, r, &req) {
		return
	}
	if err := app.svc.EmailService.SendEmail(
		r.Context(),
		req.Subject,
		req.Body,
	); err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusAccepted, w, http.StatusText(http.StatusAccepted))
}

var supportedContentTypes = map[string]string{
	"image/jpeg": "jpeg",
	"image/png":  "png",
	"image/tiff": "tiff",
	"image/webp": "webp",
	"image/avif": "avif",
	"image/heic": "heic",
}

func (app *appEnv) postSignedUpload(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var userData struct {
		Type string `json:"type"`
	}
	if !app.readJSON(w, r, &userData) {
		return
	}

	ext, ok := supportedContentTypes[userData.Type]
	if !ok {
		app.replyErr(w, r, resperr.WithUserMessagef(
			nil, "File has an unsupported content type: %q", ext,
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
	app.logStart(r)

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
	app.logStart(r)

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
	app.logStart(r)

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

	var v resperr.Validator
	v.AddIf("domain", req.Domain == "", "Can't add nothing")
	v.AddIf("domain", req.Domain == "spotlightpa.org", "Can't change spotlightpa.org!")
	if err := v.Err(); err != nil {
		app.replyErr(w, r, err)
		return
	}

	var roles []string
	if !req.Remove {
		roles = []string{"editor"}
	}

	if _, err := app.svc.Queries.UpsertRolesForDomain(
		r.Context(),
		db.UpsertRolesForDomainParams{
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
	app.logStart(r)

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
	app.logStart(r)

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
		app.replyErr(w, r, resperr.WithUserMessagef(nil,
			"Invalid email address: %q", req.Address))
		return
	}

	var roles []string
	if !req.Remove {
		roles = []string{"editor"}
	}

	if _, err := app.svc.Queries.UpsertRolesForAddress(
		r.Context(),
		db.UpsertRolesForAddressParams{
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

func (app *appEnv) listImages(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var page int32
	_ = intFromQuery(r, "page", &page)
	if page < 0 {
		app.replyErr(w, r, resperr.WithUserMessage(nil, "Invalid page"))
		return
	}

	pager := paginate.PageNumber(page)
	pager.PageSize = 100
	images, err := paginate.List(
		pager,
		r.Context(),
		app.svc.Queries.ListImages,
		db.ListImagesParams{
			Offset: pager.Offset(),
			Limit:  pager.Limit(),
		})
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	waitingFor, err := app.svc.Queries.ListImageWhereNotUploaded(r.Context())
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, struct {
		Images           []db.Image `json:"images"`
		NextPage         int32      `json:"next_page,string,omitempty"`
		WaitingForUpload bool       `json:"waiting_for_upload"`
	}{
		Images:           images,
		NextPage:         pager.NextPage,
		WaitingForUpload: len(waitingFor) != 0,
	})
}

func (app *appEnv) listAllTopics(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

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
	app.logStart(r)

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
	app.logStart(r)

	var page int32
	_ = intFromQuery(r, "page", &page)
	if page < 0 {
		app.replyErr(w, r, resperr.WithUserMessage(nil, "Invalid page"))
		return
	}

	pager := paginate.PageNumber(page)
	pager.PageSize = 100
	files, err := paginate.List(
		pager,
		r.Context(),
		app.svc.Queries.ListFiles,
		db.ListFilesParams{
			Offset: pager.Offset(),
			Limit:  pager.Limit(),
		})
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, struct {
		Files    []db.File `json:"files"`
		NextPage int32     `json:"next_page,string,omitempty"`
	}{
		Files:    files,
		NextPage: pager.NextPage,
	})
}

func (app *appEnv) postFileCreate(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

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
	app.logStart(r)

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

func (app *appEnv) listPages(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var page int32
	_ = intFromQuery(r, "page", &page)
	if page < 0 {
		app.replyErr(w, r, resperr.WithUserMessage(nil, "Invalid page"))
		return
	}

	prefix := r.URL.Query().Get("path")

	var (
		resp struct {
			Pages    []db.ListPagesRow `json:"pages"`
			NextPage int32             `json:"next_page,string,omitempty"`
		}
		err error
	)
	pager := paginate.PageNumber(page)
	pager.PageSize = 100
	resp.Pages, err = paginate.List(pager, r.Context(),
		app.svc.Queries.ListPages,
		db.ListPagesParams{
			FilePath: prefix + "%",
			Limit:    pager.Limit(),
			Offset:   pager.Offset(),
		})
	resp.NextPage = pager.NextPage
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, &resp)
}

func (app *appEnv) getPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	by := q.Get("by")
	val := q.Get("value")
	refresh, _ := boolFromQuery(r, "refresh_content_store")
	var id int64

	app.logStart(r, "by", by, "value", val, "refresh_content_store", refresh)

	var v resperr.Validator
	v.AddIf("by",
		!slices.Contains([]string{"id", "filepath", "urlpath"}, by),
		"Must specify look up method")
	v.AddIf("value",
		by == "filepath" && strings.Contains(val, "%"),
		"Filepath contains forbidden character.")
	v.AddIf("value",
		by == "id" && !intFromQuery(r, "value", &id),
		"ID must be integer")
	if err := v.Err(); err != nil {
		app.replyErr(w, r, err)
		return
	}
	var (
		page db.Page
		err  error
	)
	switch by {
	case "id":
		page, err = app.svc.Queries.GetPageByID(r.Context(), id)
	case "filepath":
		page, err = app.svc.Queries.GetPageByFilePath(r.Context(), val)
	case "urlpath":
		page, err = app.svc.Queries.GetPageByURLPath(r.Context(), val)
	default:
		err = errors.New("unreachable")
	}
	if err != nil {
		err = db.NoRowsAs404(err, "could not find page %v", q)
		app.replyErr(w, r, err)
		return
	}

	if refresh {
		if warning := app.svc.RefreshPageFromContentStore(r.Context(), &page); warning != nil {
			app.logErr(r.Context(), warning)
		}
	}
	if slices.Contains(q["select"], "-body") {
		page.Body = ""
		delete(page.Frontmatter, "raw-content")
	}
	app.replyJSON(http.StatusOK, w, page)
}

func (app *appEnv) postPage(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var userUpdate db.UpdatePageParams
	if !app.readJSON(w, r, &userUpdate) {
		return
	}

	oldPage, err := app.svc.Queries.GetPageByFilePath(r.Context(), userUpdate.FilePath)
	if err != nil {
		err = fmt.Errorf("postPage connection problem: %w", err)
		app.replyErr(w, r, err)
		return
	}

	res, err := app.svc.Queries.UpdatePage(r.Context(), userUpdate)
	if err != nil {
		err = fmt.Errorf("postPage update problem: %w", err)
		app.replyErr(w, r, err)
		return
	}
	shouldPublish := res.ShouldPublish()
	shouldNotify := res.ShouldNotify(&oldPage)
	if shouldNotify {
		if err = app.svc.Notify(r.Context(), &res, shouldPublish); err != nil {
			app.logErr(r.Context(), err)
		}
	}
	if shouldPublish {
		err, warning := app.svc.PublishPage(r.Context(), app.svc.Queries, &res)
		if warning != nil {
			app.logErr(r.Context(), warning)
		}
		if err != nil {
			err = fmt.Errorf("postPage publish problem: %w", err)
			app.replyErr(w, r, err)
			return
		}
	}
	app.replyJSON(http.StatusOK, w, &res)
}

func (app *appEnv) postPageRefresh(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	var req struct {
		ID              int64 `json:"id,string"`
		RefreshMetadata bool  `json:"refresh_metadata"`
	}
	if !app.readJSON(w, r, &req) {
		return
	}

	id := req.ID

	page, err := app.svc.Queries.GetPageByID(r.Context(), id)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find page ID %d", id)
		app.replyErr(w, r, err)
		return
	}

	if page.SourceType == "mailchimp" {
		err = app.svc.RefreshPageFromMailchimp(r.Context(), &page)
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		app.replyJSON(http.StatusOK, w, &page)
		return
	}

	if page.SourceType != "arc" {
		app.replyNewErr(http.StatusConflict, w, r,
			"cannot refresh page %d type %q", id, page.SourceType)
		return
	}

	arcID := page.SourceID
	if arcID == "" {
		app.replyNewErr(http.StatusConflict, w, r, "no arc-id on page %d", id)
		return
	}

	if fatal, err := app.svc.RefreshArcFromFeed(r.Context()); err != nil {
		if fatal {
			app.replyErr(w, r, err)
			return
		}
		app.logErr(r.Context(), err)
	}

	story, err := app.svc.Queries.GetArcByArcID(r.Context(), arcID)
	if err != nil {
		if db.IsNotFound(err) {
			err = fmt.Errorf("page %d refers to bad arc-id %q: %w", id, arcID, err)
		}
		app.replyErr(w, r, err)
		return
	}

	if warnings, err := app.svc.RefreshPageFromArcStory(r.Context(), &page, &story, req.RefreshMetadata); err != nil {
		app.replyErr(w, r, err)
		return
	} else {
		for _, w := range warnings {
			app.logErr(r.Context(), fmt.Errorf("got warning: %s", w))
		}
	}
	app.replyJSON(http.StatusOK, w, page)
}

func (app *appEnv) postPageCreate(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var req struct {
		SharedArticleID int64  `json:"shared_article_id,string"`
		PageKind        string `json:"page_kind"`
	}
	if !app.readJSON(w, r, &req) {
		return
	}
	if !slices.Contains([]string{"news", "statecollege"}, req.PageKind) {
		app.replyErr(w, r, resperr.WithUserMessage(nil, "Invalid page_kind"))
		return
	}

	sharedArt, err := app.svc.Queries.GetSharedArticleByID(r.Context(), req.SharedArticleID)
	if err != nil {
		err = db.NoRowsAs404(err, "missing id=%q", req.SharedArticleID)
		app.replyErr(w, r, err)
		return
	}

	if sharedArt.PageID.Status == pgtype.Present {
		app.replyErr(w, r, fmt.Errorf(
			"can't create new page for %d; page %d already exists",
			req.SharedArticleID, sharedArt.PageID.Int))
		return
	}

	warnings, err := app.svc.CreatePageFromArcSource(r.Context(), &sharedArt, req.PageKind)
	for _, w := range warnings {
		app.logErr(r.Context(), fmt.Errorf("got warning: %s", w))
	}
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, sharedArt)
}

func (app *appEnv) listAllPages(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	type response struct {
		Pages []db.ListAllPagesRow `json:"pages"`
	}
	var (
		res response
		err error
	)

	if res.Pages, err = app.svc.Queries.ListAllPages(r.Context()); err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, res)
}

func (app *appEnv) getSiteData(loc string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.logStart(r, "location", loc)

		type response struct {
			Configs []db.SiteDatum `json:"configs"`
		}
		var (
			res response
			err error
		)
		res.Configs, err = app.svc.Queries.GetSiteData(r.Context(), loc)
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		app.replyJSON(http.StatusOK, w, res)
	}
}

func (app *appEnv) setSiteData(loc string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.logStart(r, "location", loc)

		var req struct {
			Configs []almanack.ScheduledSiteConfig `json:"configs"`
		}
		if !app.readJSON(w, r, &req) {
			return
		}
		if len(req.Configs) < 1 {
			app.replyErr(w, r, resperr.WithUserMessage(
				nil, "No schedulable items provided"))
			return
		}

		var (
			res struct {
				Configs []db.SiteDatum `json:"configs"`
			}
			err error
		)
		res.Configs, err = app.svc.UpdateSiteConfig(r.Context(), loc, req.Configs)
		if err != nil {
			app.replyErr(w, r, err)
			return
		}

		app.replyJSON(http.StatusOK, w, res)
	}
}

func (app *appEnv) listPagesByFTS(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := q.Get("query")
	app.logStart(r, "query", query)

	var (
		pages []db.Page
		err   error
	)
	if query == "" {
		pages, err = app.svc.Queries.ListPagesByPublished(r.Context(), db.ListPagesByPublishedParams{
			Limit:  20,
			Offset: 0,
		})
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
	} else {
		pages, err = app.svc.Queries.ListPagesByFTS(r.Context(), db.ListPagesByFTSParams{
			Query: query,
			Limit: 20,
		})
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		if !strings.Contains(query, " ") {
			idpages, err := app.svc.Queries.ListPagesByInternalID(r.Context(), db.ListPagesByInternalIDParams{
				Query: fmt.Sprintf("%s:*", query),
				Limit: 5,
			})
			if err != nil {
				app.replyErr(w, r, err)
				return
			}
			pages = append(idpages, pages...)
			if len(pages) > 20 {
				pages = pages[:20]
			}
		}
	}

	if slices.Contains(q["select"], "-body") {
		for i := range pages {
			page := &pages[i]
			page.Body = ""
			delete(page.Frontmatter, "raw-content")
		}
	}
	app.replyJSON(http.StatusOK, w, pages)
}

func (app *appEnv) listArcByLastUpdated(w http.ResponseWriter, r *http.Request) {
	var page int32
	_ = intFromQuery(r, "page", &page)
	refresh, _ := boolFromQuery(r, "refresh")
	app.logStart(r, "page", page, "refresh", refresh)

	if refresh {
		if fatal, err := app.svc.RefreshArcFromFeed(r.Context()); err != nil {
			if fatal {
				app.replyErr(w, r, err)
				return
			}
			app.logErr(r.Context(), err)
		}
	}

	var (
		resp struct {
			Stories  []db.ListArcByLastUpdatedRow `json:"stories"`
			NextPage int32                        `json:"next_page,string,omitempty"`
		}
		err error
	)
	pager := paginate.PageNumber(page)
	pager.PageSize = 20
	resp.Stories, err = paginate.List(pager, r.Context(),
		app.svc.Queries.ListArcByLastUpdated,
		db.ListArcByLastUpdatedParams{
			Limit:  pager.Limit(),
			Offset: pager.Offset(),
		})
	resp.NextPage = pager.NextPage
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, &resp)
}

func (app *appEnv) postSharedArticle(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var req db.UpdateSharedArticleParams
	if !app.readJSON(w, r, &req) {
		return
	}

	article, err := app.svc.Queries.UpdateSharedArticle(r.Context(), req)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, &article)
}

func (app *appEnv) postSharedArticleFromArc(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var req struct {
		ArcID string `json:"arc_id"`
	}
	if !app.readJSON(w, r, &req) {
		return
	}
	if fatal, err := app.svc.RefreshArcFromFeed(r.Context()); err != nil {
		if fatal {
			app.replyErr(w, r, err)
			return
		}
		app.logErr(r.Context(), err)
	}

	article, err := app.svc.Queries.UpsertSharedArticleFromArc(r.Context(), req.ArcID)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, &article)
}
