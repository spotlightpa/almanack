package api

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/carlmjohnson/flowmatic"
	"github.com/earthboundkid/emailx/v2"
	"github.com/earthboundkid/resperr/v2"
	"github.com/jackc/pgx/v5"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/gdocs"
	"github.com/spotlightpa/almanack/internal/google"
	"github.com/spotlightpa/almanack/internal/paginate"
	"github.com/spotlightpa/almanack/internal/slicex"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
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
	ctx := context.WithoutCancel(r.Context())
	if err := app.svc.EmailService.SendEmail(
		ctx,
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
		app.replyErr(w, r, error(resperr.E{M: fmt.Sprintf("File has an unsupported content type: %q", ext)}))
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
	if dbImage, err := app.svc.Queries.UpsertImage(r.Context(), db.UpsertImageParams{
		Path: res.FileName,
		Type: ext,
	}); err != nil {
		app.replyErr(w, r, err)
		return
	} else if dbImage.IsUploaded {
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
		app.replyErr(w, r, resperr.E{M: fmt.Sprintf("Invalid email address: %q", req.Address)})
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
		app.replyErr(w, r, resperr.E{M: "Invalid page"})
		return
	}

	pager := paginate.PageNumber(page)
	pager.PageSize = 100
	query := r.URL.Query().Get("query")

	var (
		images []db.Image
		err    error
	)

	if query != "" {
		images, err = paginate.List(
			pager,
			r.Context(),
			app.svc.Queries.ListImagesByFTS,
			db.ListImagesByFTSParams{
				Limit:  pager.Limit(),
				Offset: pager.Offset(),
				Query:  query,
			})
	} else {
		images, err = paginate.List(
			pager,
			r.Context(),
			app.svc.Queries.ListImages,
			db.ListImagesParams{
				Offset: pager.Offset(),
				Limit:  pager.Limit(),
			})
	}
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
		app.replyErr(w, r, resperr.E{M: "Invalid page"})
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
		app.replyErr(w, r, resperr.E{M: "Invalid page"})
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

	oldPage, err := app.svc.Queries.GetPageByID(r.Context(), userUpdate.ID)
	if err != nil {
		err = fmt.Errorf("postPage connection problem: %w", err)
		app.replyErr(w, r, err)
		return
	}
	ctx := context.WithoutCancel(r.Context())
	res, err := app.svc.Queries.UpdatePage(ctx, userUpdate)
	if err != nil {
		err = fmt.Errorf("postPage update problem: %w", err)
		app.replyErr(w, r, err)
		return
	}
	shouldPublish := res.ShouldPublish()
	shouldNotify := res.ShouldNotify(&oldPage)
	if shouldPublish {
		err = app.svc.Tx.Begin(ctx, pgx.TxOptions{}, func(txq *db.Queries) (txerr error) {
			err, warning := app.svc.PublishPage(ctx, txq, &res)
			if warning != nil {
				app.logErr(r.Context(), warning)
			}
			return err
		})
		if err != nil {
			err = fmt.Errorf("postPage publish problem: %w", err)
			app.replyErr(w, r, err)
			return
		}
	}
	if shouldNotify {
		if err = app.svc.Notify(ctx, &res, shouldPublish); err != nil {
			app.logErr(ctx, err)
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

	switch page.SourceType {
	case "gdocs":
		dbDoc, err := app.svc.Queries.GetGDocsByExternalIDWhereProcessed(r.Context(), page.SourceID)
		if err != nil {
			app.replyErr(w, r, err)
			return
		}
		if req.RefreshMetadata {
			if page.Frontmatter == nil {
				page.Frontmatter = make(db.Map)
			}
			fm := map[string]any{
				"byline":            dbDoc.Metadata.Byline,
				"authors":           stringx.ExtractNames(dbDoc.Metadata.Byline),
				"title":             dbDoc.Metadata.Hed,
				"description":       dbDoc.Metadata.Description,
				"image":             dbDoc.Metadata.LedeImage,
				"image-credit":      dbDoc.Metadata.LedeImageCredit,
				"image-description": dbDoc.Metadata.LedeImageDescription,
				"image-caption":     dbDoc.Metadata.LedeImageCaption,
				// Fields not exposed to Shared Admin
				"kicker": dbDoc.Metadata.Eyebrow,
				"blurb": cmp.Or(
					dbDoc.Metadata.Blurb,
					dbDoc.Metadata.Description,
				),
				"linktitle":     dbDoc.Metadata.LinkTitle,
				"title-tag":     dbDoc.Metadata.SEOTitle,
				"og-title":      dbDoc.Metadata.OGTitle,
				"twitter-title": dbDoc.Metadata.TwitterTitle,
				"layout":        dbDoc.Metadata.Layout,
			}
			// Remove blanks
			maps.DeleteFunc(fm, func(key string, v any) bool {
				if s, ok := v.(string); ok {
					return s == ""
				}
				return false
			})
			maps.Copy(page.Frontmatter, fm)
		}
		page.Body = dbDoc.ArticleMarkdown
		app.replyJSON(http.StatusOK, w, &page)
		return
	case "mailchimp":
		app.replyNewErr(http.StatusConflict, w, r, "can not refresh source-type mailchimp; id=%d", id)
		return

	case "arc":
		app.replyNewErr(http.StatusConflict, w, r, "can not refresh source-type arc; id=%d", id)
		return
	default:
		app.replyNewErr(http.StatusConflict, w, r,
			"cannot refresh page %d type %q", id, page.SourceType)
		return
	}
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
	if !slices.Contains([]string{"news", "statecollege", "berks"}, req.PageKind) {
		app.replyErr(w, r, resperr.E{M: "Invalid page_kind"})
		return
	}

	sharedArt, err := app.svc.Queries.GetSharedArticleByID(r.Context(), req.SharedArticleID)
	if err != nil {
		err = db.NoRowsAs404(err, "missing id=%q", req.SharedArticleID)
		app.replyErr(w, r, err)
		return
	}

	if sharedArt.PageID.Valid {
		app.replyNewErr(http.StatusConflict, w, r,
			"can't create new page for %d; page %d already exists",
			req.SharedArticleID, sharedArt.PageID.Int64)
		return
	}

	switch sharedArt.SourceType {
	case "gdocs":
		err = app.svc.CreatePageFromGDocsDoc(r.Context(), &sharedArt, req.PageKind)
		if err != nil {
			app.replyErr(w, r, err)
			return
		}

		app.replyJSON(http.StatusOK, w, sharedArt)
		return

	default:
		app.replyNewErr(http.StatusConflict, w, r,
			"can't create new page for %d; bad source_type: %q",
			req.SharedArticleID, sharedArt.SourceType)
		return
	}
}

func (app *appEnv) siteDataGet(loc string) http.HandlerFunc {
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

func (app *appEnv) getSiteData(w http.ResponseWriter, r *http.Request) http.Handler {
	loc := r.URL.Query().Get("location")
	return app.siteDataGet(loc)
}

func (app *appEnv) siteDataSet(loc string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.logStart(r, "location", loc)

		var req struct {
			Configs []almanack.ScheduledSiteConfig `json:"configs"`
		}
		if !app.readJSON(w, r, &req) {
			return
		}
		if len(req.Configs) < 1 {
			app.replyErr(w, r, resperr.E{M: "No schedulable items provided"})
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

func (app *appEnv) postSiteData(w http.ResponseWriter, r *http.Request) http.Handler {
	loc := r.URL.Query().Get("location")
	return app.siteDataSet(loc)
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
		switch {
		case strings.HasPrefix(query, "http"):
			u, err := url.Parse(query)
			if err != nil {
				l := almlog.FromContext(r.Context())
				l.DebugContext(r.Context(), "listPagesByFTS: bad URL", "query", query)
				break
			}
			pathPage, err := app.svc.Queries.GetPageByURLPath(r.Context(), u.Path)
			if db.IsNotFound(err) {
				break
			}
			if err != nil {
				app.replyErr(w, r, err)
				return
			}
			pathPages := []db.Page{pathPage}
			pages = append(pathPages, pages...)
			pages = pages[:min(20, len(pages))]

		case !strings.ContainsAny(query, " /:*"):
			idpages, err := app.svc.Queries.ListPagesByInternalID(r.Context(), db.ListPagesByInternalIDParams{
				Query: fmt.Sprintf("%s:*", query),
				Limit: 5,
			})
			if err != nil {
				app.replyErr(w, r, err)
				return
			}
			pages = slices.Concat(idpages, pages)
			slicex.UniquesFunc(&pages, func(p db.Page) int64 {
				return p.ID
			})
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

func (app *appEnv) postSharedArticleFromGDocs(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	l := almlog.FromContext(r.Context())
	var req struct {
		ID              string `json:"external_gdocs_id"`
		ForceUpdate     bool   `json:"force_update"`
		RefreshMetadata bool   `json:"refresh_metadata"`
	}
	if !app.readJSON(w, r, &req) {
		return
	}

	id, err := gdocs.NormalizeID(req.ID)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	dbDoc, err := app.svc.Queries.GetGDocsByExternalIDWhereProcessed(r.Context(), id)
	if err != nil {
		err = db.NoRowsAs404(err, "missing external_gdocs_id=%q", req.ID)
		app.replyErr(w, r, err)
		return
	}

	if !req.ForceUpdate {
		l.Debug("postSharedArticleFromGDocs", "force_update", false)
		art, err := app.svc.Queries.GetSharedArticleBySource(r.Context(), db.GetSharedArticleBySourceParams{
			SourceType: "gdocs",
			SourceID:   dbDoc.ExternalID,
		})
		switch {
		// Skip update if it exists
		case err == nil:
			l.Debug("postSharedArticleFromGDocs: skipping")
			app.replyJSON(http.StatusOK, w, art)
			return
		case db.IsNotFound(err):
			break
		case err != nil:
			app.replyErr(w, r, err)
			return
		}
	}

	art, err := app.svc.UpsertSharedArticleForGDoc(r.Context(), &dbDoc, req.RefreshMetadata)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, art)
}

func (app *appEnv) postGDocsDoc(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	l := almlog.FromContext(r.Context())

	var req struct {
		ID string `json:"external_gdocs_id"`
	}
	if !app.readJSON(w, r, &req) {
		return
	}
	id, err := gdocs.NormalizeID(req.ID)
	if err != nil {
		l.Warn("postSharedArticleFromGDocs: bad id", "id", req.ID)
		app.replyErr(w, r, err)
		return
	}

	art, err := app.svc.CreateGDocsDoc(r.Context(), id)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, art)
}

func (app *appEnv) getGDocsDoc(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var id int64
	if !intFromQuery(r, "id", &id) {
		app.replyNewErr(http.StatusBadRequest, w, r, "missing ID")
		return
	}

	dbDoc, err := app.svc.Queries.GetGDocsByID(r.Context(), id)
	if err != nil {
		err = db.NoRowsAs404(err, "missing g_docs_doc.id=%d", id)
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, dbDoc)
}

func (app *appEnv) postDonorWall(w http.ResponseWriter, r *http.Request) http.Handler {
	app.logStart(r)
	sheetID, err := app.svc.Queries.GetOption(r.Context(), "donor-wall")
	if err != nil {
		return app.jsonErr(err)
	}
	if err := app.svc.ConfigureGoogleCert(r.Context()); err != nil {
		return app.jsonErr(err)
	}
	cl, err := app.svc.Gsvc.SheetsClient(r.Context())
	if err != nil {
		return app.jsonErr(err)
	}
	files, err := google.SheetToDonorWall(r.Context(), cl, sheetID)
	if err != nil {
		return app.jsonErr(err)
	}
	paths := make([]string, 0, len(files))
	for fpath := range files {
		paths = append(paths, fpath)
	}
	err = flowmatic.Each(5, paths, func(fpath string) error {
		msg := fmt.Sprintf("%s: updating donor wall", path.Base(fpath))
		obj := files[fpath]
		content, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return err
		}
		return app.svc.ContentStore.UpdateFile(r.Context(), msg, fpath, content)
	})
	if err != nil {
		return app.jsonErr(err)
	}
	return app.jsonOK("OK")
}

func (app *appEnv) postPageLoad(w http.ResponseWriter, r *http.Request) http.Handler {
	// Load a page already published in the content store and add it to the database.
	// Steps:
	// - Get the path from the request
	// - Load it from the store
	// - Shove it in the DB
	app.logStart(r)
	if err := r.ParseForm(); err != nil {
		return app.jsonErr(err)
	}
	path := r.PostForm.Get("path")
	var v resperr.Validator
	v.AddIf("path", path == "", "missing path")
	if err := v.Err(); err != nil {
		return app.jsonErr(err)
	}
	content, err := app.svc.ContentStore.GetFile(r.Context(), path)
	if err != nil {
		return app.jsonErr(err)
	}
	if _, err := db.CreatePageFromContent(r.Context(), app.svc.Tx, path, content); err != nil {
		return app.jsonErr(err)
	}
	return app.jsonOK("ok")
}
