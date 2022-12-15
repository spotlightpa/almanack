package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/carlmjohnson/emailx"
	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/internal/arc"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/mailchimp"
	"github.com/spotlightpa/almanack/internal/paginate"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"golang.org/x/exp/slices"
)

func (app *appEnv) listAllArcStories(w http.ResponseWriter, r *http.Request) {
	var page int32
	mustIntParam(r, "page", &page)
	app.Printf("start listAllArcStories page %d", page)

	var (
		resp struct {
			Contents []almanack.ArcStory `json:"contents"`
			NextPage int32               `json:"next_page,string,omitempty"`
		}
		err error
	)
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
		NextPage int32               `json:"next_page,string,omitempty"`
	}
	var (
		feed *arc.API
		err  error
	)
	if feed, err = app.svc.FetchArcFeed(r.Context()); err != nil {
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
			feed *arc.API
			err  error
		)
		if feed, err = app.svc.FetchArcFeed(r.Context()); err != nil {
			app.replyErr(w, r, err)
			return
		}
		if err := app.svc.StoreFeed(r.Context(), feed); err != nil {
			app.replyErr(w, r, err)
			return
		}
		for i := range feed.Contents {
			if feed.Contents[i].ID == userData.ID {
				story.FeedItem = feed.Contents[i]
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
	app.Printf("start postSignedUpload")
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
		app.replyErr(w, r, resperr.WithUserMessagef(nil,
			"Invalid email address: %q", req.Address))
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

func (app *appEnv) listImages(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listImages")

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

	app.replyJSON(http.StatusOK, w, struct {
		Images   []db.Image `json:"images"`
		NextPage int32      `json:"next_page,string,omitempty"`
	}{
		Images:   images,
		NextPage: pager.NextPage,
	})
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

func (app *appEnv) listPages(w http.ResponseWriter, r *http.Request) {
	app.Printf("start listPages")

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
	var id int64
	mustIntParam(r, "id", &id)
	app.Printf("start getPage for %d", id)
	page, err := app.svc.Queries.GetPageByID(r.Context(), id)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find page ID %d", id)
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, page)
}

func (app *appEnv) getPageByFilePath(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	path := q.Get("path")
	app.Printf("start getPageByFilePath for %q", path)
	page, err := app.svc.Queries.GetPageByFilePath(r.Context(), path)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find page %q", path)
		app.replyErr(w, r, err)
		return
	}
	if slices.Contains(q["select"], "-body") {
		page.Body = ""
		delete(page.Frontmatter, "raw-content")
	}
	app.replyJSON(http.StatusOK, w, page)
}

func (app *appEnv) getPageByURLPath(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	path := q.Get("path")
	app.Printf("starting getPageByURLPath for %q", path)
	var v resperr.Validator
	v.AddIf("path", strings.Contains(path, "%"), "Contains forbidden character.")
	if err := v.Err(); err != nil {
		app.replyErr(w, r, err)
		return
	}
	page, err := app.svc.Queries.GetPageByURLPath(r.Context(), path)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find page %q", path)
		app.replyErr(w, r, err)
		return
	}
	if slices.Contains(q["select"], "-body") {
		page.Body = ""
		delete(page.Frontmatter, "raw-content")
	}
	app.replyJSON(http.StatusOK, w, page)

}

func (app *appEnv) getPageWithContent(w http.ResponseWriter, r *http.Request) {
	var id int64
	mustIntParam(r, "id", &id)
	app.Printf("start getPage for %d", id)
	page, err := app.svc.Queries.GetPageByID(r.Context(), id)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find page ID %d", id)
		app.replyErr(w, r, err)
		return
	}
	if warning := app.svc.RefreshPageFromContentStore(r.Context(), &page); warning != nil {
		app.logErr(r.Context(), warning)
	}
	app.replyJSON(http.StatusOK, w, page)
}

func (app *appEnv) postPage(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postPage")

	var userUpdate db.UpdatePageParams
	if !app.readJSON(w, r, &userUpdate) {
		return
	}

	oldPage, err := app.svc.Queries.GetPageByFilePath(r.Context(), userUpdate.FilePath)
	if err != nil {
		errutil.Prefix(&err, "postPage connection problem")
		app.replyErr(w, r, err)
		return
	}

	res, err := app.svc.Queries.UpdatePage(r.Context(), userUpdate)
	if err != nil {
		errutil.Prefix(&err, "postPage update problem")
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
			errutil.Prefix(&err, "postPage publish problem")
			app.replyErr(w, r, err)
			return
		}
	}
	app.replyJSON(http.StatusOK, w, &res)
}

func (app *appEnv) getPageForArcID(w http.ResponseWriter, r *http.Request) {
	arcID := r.URL.Query().Get("arc_id")
	app.Printf("start getPageForArcID for %q", arcID)

	dbArt, err := app.svc.Queries.GetArticleByArcID(r.Context(), arcID)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find Arc ID %q", arcID)
		app.replyErr(w, r, err)
		return
	}
	filepath := dbArt.SpotlightPAPath.String
	if filepath == "" {
		// Empty string signals to bring up creation page
		app.replyJSON(http.StatusOK, w, "")
		return
	}
	app.pageIDForFilepath(w, r, filepath)
}

func (app *appEnv) pageIDForFilepath(w http.ResponseWriter, r *http.Request, filepath string) {
	page, err := app.svc.Queries.GetPageByFilePath(r.Context(), filepath)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, page.ID)
}

func (app *appEnv) postPageForArcID(w http.ResponseWriter, r *http.Request) {
	app.Print("start postPageForArcID")

	var req struct {
		ArcID    string `json:"arc_id"`
		PageKind string `json:"page_kind"`
	}
	if !app.readJSON(w, r, &req) {
		return
	}
	if !slices.Contains([]string{"news", "statecollege"}, req.PageKind) {
		app.replyErr(w, r, resperr.WithUserMessage(nil, "Invalid page_kind"))
		return
	}

	app.Printf("saving page for %q", req.ArcID)
	dbArt, err := app.svc.Queries.GetArticleByArcID(r.Context(), req.ArcID)
	if err != nil {
		err = db.NoRowsAs404(err, "could not find Arc ID %q", req.ArcID)
		app.replyErr(w, r, err)
		return
	}
	filepath := dbArt.SpotlightPAPath.String
	if filepath != "" {
		app.logErr(r.Context(), fmt.Errorf(
			"can't create new page for %q; page %q already exists",
			req.ArcID, filepath))
		app.pageIDForFilepath(w, r, filepath)
		return
	}

	page, err := app.svc.PageFromArcArticle(r.Context(), &dbArt, req.PageKind)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	app.replyJSON(http.StatusOK, w, page.ID)
}

func (app *appEnv) postPageRefresh(w http.ResponseWriter, r *http.Request) {
	app.Print("start postPageRefresh")
	var req struct {
		ID int64 `json:"id,string"`
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

	// TODO: test if it's a MailChimp page
	arcID, _ := page.Frontmatter["arc-id"].(string)
	if arcID == "" {
		app.replyNewErr(http.StatusConflict, w, r, "no arc-id on page %d", id)
		return
	}

	feed, feedErr := app.svc.FetchArcFeed(r.Context())
	if feedErr != nil {
		// Keep trucking even if you can't load feed
		app.logErr(r.Context(), feedErr)
	}
	if feedErr == nil {
		if err = app.svc.StoreFeed(r.Context(), feed); err != nil {
			app.replyErr(w, r, err)
			return
		}
	}

	story, err := app.svc.Queries.GetArcByArcID(r.Context(), arcID)
	if err != nil {
		if db.IsNotFound(err) {
			err = fmt.Errorf("page %d refers to bad arc-id %q: %w", id, arcID, err)
		}
		app.replyErr(w, r, err)
		return
	}

	if err = app.svc.RefreshPageFromArcStory(r.Context(), &page, &story); err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, page)
}

func (app *appEnv) listAllPages(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listSpotlightPAArticles")
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
		app.Printf("starting getSiteData(%q)", loc)

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
		app.Printf("starting setSiteData(%q)", loc)

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

func (app *appEnv) postRefreshPageFromMailchimp(w http.ResponseWriter, r *http.Request) {
	var id int64
	mustIntParam(r, "id", &id)
	app.Printf("start postRefreshPageFromMailchimp for %d", id)

	archiveURL, err := app.svc.Queries.GetArchiveURLForPageID(r.Context(), id)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	if archiveURL == "" {
		app.replyNewErr(http.StatusConflict, w, r, "no archiveURL for page %d", id)
	}
	body, err := mailchimp.ImportPage(r.Context(), app.svc.Client, archiveURL)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	page, err := app.svc.Queries.UpdatePageRawContent(r.Context(), db.UpdatePageRawContentParams{
		ID:         id,
		RawContent: body,
	})
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	app.replyJSON(http.StatusOK, w, &page)
}

func (app *appEnv) listPagesByFTS(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := q.Get("query")
	app.Printf("start getPageByFTS for %q", query)

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
