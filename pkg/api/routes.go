package api

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"golang.org/x/net/context/ctxhttp"

	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/httpjson"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func (app *appEnv) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: app.Logger}))
	r.Use(app.versionMiddleware)
	r.Get("/api/healthcheck", app.ping)
	r.Get(`/api/healthcheck/{code:\d{3}}`, app.pingErr)
	r.Get(`/api/proxy-image/{encURL}`, app.getProxyImage)
	r.Get(`/api/cron`, app.getCron)
	r.Route("/api", func(r chi.Router) {
		r.Use(app.authMiddleware)
		r.Get("/user-info", app.userInfo)
		r.Get("/bookmarklet/{slug}", app.getBookmarklet)
		r.With(
			app.hasRoleMiddleware("editor"),
		).Group(func(r chi.Router) {
			r.Get("/available-articles", app.listAvailableArcStories)
			r.Get(`/list-available/{page:\d+}`, app.listAvailableArcStories)
			r.Get("/available-articles/{id}", app.getArcStory)
			r.Get("/mailchimp-signup-url", app.getSignupURL)
		})
		r.With(
			app.hasRoleMiddleware("Spotlight PA"),
		).Group(func(r chi.Router) {
			r.Get("/upcoming-articles", app.listAllArcStories)
			r.Get(`/list-any-arc/{page:\d+}`, app.listAllArcStories)
			r.Get("/list-arc-refresh", app.listWithArcRefresh)
			r.Post("/available-articles", app.postAlmanackArcStory)
			r.Post("/message", app.postMessage)
			r.Get("/scheduled-articles/{id}", app.getScheduledArticle)
			r.Post("/scheduled-articles", app.postScheduledArticle)
			r.Post("/create-signed-upload", app.postSignedUpload)
			r.Post("/image-update", app.postImageUpdate)
			r.Get("/authorized-domains", app.listDomains)
			r.Post("/authorized-domains", app.postDomain)
			r.Get("/spotlightpa-articles", app.listSpotlightPAArticles)
			r.Get("/images", app.listImages)
			r.Get("/editors-picks", app.getEditorsPicks)
			r.Post("/editors-picks", app.postEditorsPicks)
			r.Get("/all-topics", app.listAllTopics)
			r.Get("/all-series", app.listAllSeries)
			r.Get("/files-list", app.listFiles)
			r.Post("/files-create", app.postFileCreate)
			r.Post("/files-update", app.postFileUpdate)
		})
	})
	r.NotFound(app.notFound)

	return r
}

func (app *appEnv) notFound(w http.ResponseWriter, r *http.Request) {
	app.errorResponse(w, r, resperr.NotFound(r))
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.Println("start ping")
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	w.Write(b)
}

func (app *appEnv) pingErr(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	statusCode, _ := strconv.Atoi(code)
	app.Printf("start pingErr %q", code)

	app.errorResponse(w, r, resperr.New(
		statusCode, "got test ping %q", code,
	))

	return
}

func (app *appEnv) userInfo(w http.ResponseWriter, r *http.Request) {
	app.Println("start userInfo")
	userinfo, err := netlifyid.FromRequest(r)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, userinfo)
}

func (app *appEnv) getProxyImage(w http.ResponseWriter, r *http.Request) {
	app.Println("start getProxyImage")

	encURL := chi.URLParam(r, "encURL")
	decURL, err := base64.URLEncoding.DecodeString(encURL)
	if err != nil {
		app.errorResponse(w, r, resperr.New(
			http.StatusBadRequest, "could not decode URL param: %w", err,
		))
		return
	}
	u := string(decURL)
	app.Printf("requested %q", u)
	inWhitelist := false
	for _, prefix := range []string{
		"https://www.inquirer.com/resizer/",
		"https://arc-anglerfish-arc2-prod-pmn.s3.amazonaws.com/public/",
	} {
		if strings.HasPrefix(u, prefix) {
			inWhitelist = true
			break
		}
	}
	if !inWhitelist {
		app.errorResponse(w, r, resperr.New(
			http.StatusBadRequest, "untrust image URL: %s", u,
		))
		return
	}
	ctype, body, err := getImage(r.Context(), app.svc.Client, u)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	w.Header().Set("Content-Type", ctype)
	ext := strings.TrimPrefix(ctype, "image/")
	disposition := fmt.Sprintf(`attachment; filename="image.%s"`, ext)
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Cache-Control", "public, max-age=900")
	if _, err = w.Write(body); err != nil {
		app.logErr(r.Context(), err)
	}
}

func getImage(ctx context.Context, c *http.Client, srcurl string) (ctype string, body []byte, err error) {
	var res *http.Response
	res, err = ctxhttp.Get(ctx, c, srcurl)
	if err != nil {
		return
	}
	defer res.Body.Close()

	const (
		megabyte = 1 << 20
		maxSize  = 25 * megabyte
	)
	buf := bufio.NewReader(http.MaxBytesReader(nil, res.Body, maxSize))
	// http.DetectContentType only uses first 512 bytes
	peek, err := buf.Peek(512)
	if err != nil && err != io.EOF {
		return "", nil, err
	}

	if ct := http.DetectContentType(peek); strings.HasPrefix(ct, "image/jpeg") {
		ctype = "image/jpeg"
	} else if strings.HasPrefix(ct, "image/png") {
		ctype = "image/png"
	} else {
		return "", nil, resperr.WithCodeAndMessage(
			fmt.Errorf("%q did not have proper MIME type", srcurl),
			http.StatusBadRequest,
			"URL must be an image",
		)
	}

	body, err = ioutil.ReadAll(buf)
	if err != nil {
		return "", nil, err
	}
	return
}

func (app *appEnv) listAllArcStories(w http.ResponseWriter, r *http.Request) {
	page := 0
	if pageStr := chi.URLParam(r, "page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err != nil {
			app.errorResponse(w, r, fmt.Errorf(
				"bad argument to listAllArcStories: %w", err,
			))
			return
		} else {
			page = parsed
		}
	}

	app.Printf("start listAllArcStories page %d", page)

	type response struct {
		Contents []almanack.ArcStory `json:"contents"`
		NextPage int                 `json:"next_page,omitempty"`
	}
	var (
		resp response
		err  error
	)
	resp.Contents, resp.NextPage, err = app.svc.ListAllArcStories(r.Context(), page)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, &resp)
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
		app.errorResponse(w, r, err)
		return
	}
	var resp response
	resp.Contents, resp.NextPage, err = app.svc.ListAllArcStories(r.Context(), 0)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, resp)
}

func (app *appEnv) postAlmanackArcStory(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting postAlmanackArcStory")

	var userData struct {
		ID         string          `json:"_id"`
		Note       string          `json:"almanack-note,omitempty"`
		Status     almanack.Status `json:"almanack-status,omitempty"`
		RefreshArc bool            `json:"almanack-refresh-arc"`
	}
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.errorResponse(w, r, err)
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
			app.errorResponse(w, r, err)
			return
		}
		if err := app.svc.StoreFeed(r.Context(), feed); err != nil {
			app.errorResponse(w, r, err)
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
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusAccepted, w, &userData)
}

func (app *appEnv) listAvailableArcStories(w http.ResponseWriter, r *http.Request) {
	page := 0
	if pageStr := chi.URLParam(r, "page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err != nil {
			app.errorResponse(w, r, fmt.Errorf(
				"bad argument to listAvailableArcStories: %w", err,
			))
			return
		} else {
			page = parsed
		}
	}
	app.Printf("starting listAvailableArcStories page %d", page)

	type response struct {
		Contents []almanack.ArcStory `json:"contents"`
		NextPage int                 `json:"next_page,string,omitempty"`
	}
	var (
		res response
		err error
	)
	if res.Contents, res.NextPage, err = app.svc.ListAvailableArcStories(
		r.Context(), page,
	); err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.jsonResponse(http.StatusOK, w, res)
}

func (app *appEnv) getArcStory(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	app.Printf("starting getArcStory %s", articleID)

	article, err := app.svc.GetArcStory(r.Context(), articleID)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	if article.Status != almanack.StatusAvailable &&
		article.Status != almanack.StatusPlanned {
		// Let Spotlight PA users get article regardless of its status
		if err := app.auth.HasRole(r, "Spotlight PA"); err != nil {
			app.errorResponse(w, r, resperr.New(
				http.StatusNotFound, "user unauthorized to view article: %w", err,
			))
			return
		}
	}

	app.jsonResponse(http.StatusOK, w, article)
}

func (app *appEnv) postMessage(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting postMessage")
	type request struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	var req request
	if err := httpjson.DecodeRequest(w, r, &req); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	if err := app.email.SendEmail(req.Subject, req.Body); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusAccepted, w, http.StatusText(http.StatusAccepted))
}

func (app *appEnv) getScheduledArticle(w http.ResponseWriter, r *http.Request) {
	articleID := chi.URLParam(r, "id")
	app.Printf("start getScheduledArticle %s", articleID)

	article, err := app.svc.GetScheduledArticle(r.Context(), articleID)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.jsonResponse(http.StatusOK, w, article)
}

func (app *appEnv) postScheduledArticle(w http.ResponseWriter, r *http.Request) {
	app.Println("start postScheduledArticle")

	var userData struct {
		almanack.SpotlightPAArticle
		RefreshArc bool `json:"almanack-refresh-arc"`
	}
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.errorResponse(w, r, err)
		return
	}

	if userData.RefreshArc {
		var (
			feed *almanack.ArcAPI
			err  error
		)
		if feed, err = app.FetchFeed(r.Context()); err != nil {
			app.errorResponse(w, r, err)
			return
		}
		if err = app.svc.StoreFeed(r.Context(), feed); err != nil {
			app.errorResponse(w, r, err)
			return
		}
		if err = app.svc.ResetSpotlightPAArticleArcData(r.Context(), &userData.SpotlightPAArticle); err != nil {
			app.errorResponse(w, r, err)
			return
		}
	}

	if err := app.svc.SaveScheduledArticle(r.Context(), &userData.SpotlightPAArticle); err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.jsonResponse(http.StatusAccepted, w, &userData.SpotlightPAArticle)
}

var supportedContentTypes = map[string]string{
	"image/jpeg": "jpeg",
	"image/png":  "png",
}

func (app *appEnv) postSignedUpload(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postSignedUpload")
	var userData struct {
		Type string `json:"type"`
	}
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	ext, ok := supportedContentTypes[userData.Type]
	if !ok {
		app.errorResponse(w, r, resperr.New(
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
	res.SignedURL, res.FileName, err = almanack.GetSignedImageUpload(app.svc.ImageStore, ext)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	if n, err := app.svc.Querier.CreateImagePlaceholder(r.Context(), db.CreateImagePlaceholderParams{
		Path: res.FileName,
		Type: ext,
	}); err != nil {
		app.errorResponse(w, r, err)
		return
	} else if n != 1 {
		// Log and continue
		app.logErr(r.Context(),
			fmt.Errorf("creating image %q but it already exists", res.FileName))
	}
	app.jsonResponse(http.StatusOK, w, &res)
}

func (app *appEnv) postImageUpdate(w http.ResponseWriter, r *http.Request) {
	app.Println("start postImageUpdate")

	var userData db.UpdateImageParams
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	var (
		res db.Image
		err error
	)
	if res, err = app.svc.Querier.UpdateImage(r.Context(), userData); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, &res)
}

func (app *appEnv) getSignupURL(w http.ResponseWriter, r *http.Request) {
	app.Println("start getSignupURL")
	app.jsonResponse(http.StatusOK, w, app.mailchimpSignupURL)
}

func (app *appEnv) listDomains(w http.ResponseWriter, r *http.Request) {
	app.Println("start listDomains")
	type response struct {
		Domains []string `json:"domains"`
	}

	domains, err := app.svc.Querier.ListDomainsWithRole(r.Context(), "editor")
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, response{
		domains,
	})
}

func (app *appEnv) postDomain(w http.ResponseWriter, r *http.Request) {
	app.Println("start postDomain")
	type request struct {
		Domain string `json:"domain"`
	}
	type response struct {
		Domains []string `json:"domains"`
	}
	var req request
	if err := httpjson.DecodeRequest(w, r, &req); err != nil {
		app.errorResponse(w, r, err)
		return
	}

	if _, err := app.svc.Querier.AppendRoleToDomain(
		r.Context(),
		db.AppendRoleToDomainParams{
			Domain: req.Domain,
			Role:   "editor",
		},
	); err != nil {
		app.errorResponse(w, r, err)
		return
	}

	domains, err := app.svc.Querier.ListDomainsWithRole(r.Context(), "editor")
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, response{
		domains,
	})
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

	if res.Articles, err = app.svc.Querier.ListSpotlightPAArticles(r.Context()); err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.jsonResponse(http.StatusOK, w, res)
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

	if res.Images, err = app.svc.Querier.ListImages(r.Context(), db.ListImagesParams{
		Offset: 0,
		Limit:  100,
	}); err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.jsonResponse(http.StatusOK, w, res)
}

func (app *appEnv) getBookmarklet(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	app.Printf("starting getBookmarklet for %q", slug)

	arcid, err := app.svc.Querier.GetArticleIDFromSlug(r.Context(), slug)
	if err != nil && !db.IsNotFound(err) {
		app.logErr(r.Context(), err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	if arcid == "" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r,
		fmt.Sprintf("/articles/%s/schedule", arcid),
		http.StatusTemporaryRedirect)
}

func (app *appEnv) getEditorsPicks(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting getEditorsPicks")
	resp, err := almanack.GetEditorsPicks(r.Context(), app.svc.Querier)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, resp)
}

func (app *appEnv) postEditorsPicks(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting postEditorsPicks")
	var req almanack.EditorsPicks
	if err := httpjson.DecodeRequest(w, r, &req); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	if err := almanack.SetEditorsPicks(
		r.Context(),
		app.svc.Querier,
		app.svc.ContentStore,
		&req,
	); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	resp, err := almanack.GetEditorsPicks(r.Context(), app.svc.Querier)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, resp)
}

func (app *appEnv) listAllTopics(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listAllTopics")
	t, err := app.svc.Querier.ListAllTopics(r.Context())
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, struct {
		Topics []string `json:"topics"`
	}{t})
}

func (app *appEnv) listAllSeries(w http.ResponseWriter, r *http.Request) {
	app.Printf("starting listAllSeries")
	s, err := app.svc.Querier.ListAllSeries(r.Context())
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, struct {
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

	if res.Files, err = app.svc.Querier.ListFiles(r.Context(), db.ListFilesParams{
		Offset: 0,
		Limit:  100,
	}); err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.jsonResponse(http.StatusOK, w, res)
}

func (app *appEnv) postFileCreate(w http.ResponseWriter, r *http.Request) {
	app.Printf("start postFileCreate")
	var userData struct {
		MimeType string `json:"mimeType"`
		FileName string `json:"filename"`
	}
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	type response struct {
		SignedURL   string `json:"signed-url"`
		FileURL     string `json:"file-url"`
		Disposition string `json:"disposition"`
	}
	var (
		res response
		err error
	)

	res.SignedURL, res.FileURL, res.Disposition, err = almanack.GetSignedFileUpload(
		app.svc.FileStore,
		userData.FileName,
	)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}
	if n, err := app.svc.Querier.CreateFilePlaceholder(r.Context(),
		db.CreateFilePlaceholderParams{
			Filename: userData.FileName,
			Type:     userData.MimeType,
			URL:      res.FileURL,
		}); err != nil {
		app.errorResponse(w, r, err)
		return
	} else if n != 1 {
		// Log and continue
		app.logErr(r.Context(),
			fmt.Errorf("creating file %q but it already exists", res.FileURL))
	}
	app.jsonResponse(http.StatusOK, w, &res)
}

func (app *appEnv) postFileUpdate(w http.ResponseWriter, r *http.Request) {
	app.Println("start postFileUpdate")

	var userData db.UpdateFileParams
	if err := httpjson.DecodeRequest(w, r, &userData); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	var (
		res db.File
		err error
	)
	if res, err = app.svc.Querier.UpdateFile(r.Context(), userData); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, &res)
}

func (app *appEnv) getCron(w http.ResponseWriter, r *http.Request) {
	if err := app.svc.PopScheduledArticles(r.Context()); err != nil {
		app.errorResponse(w, r, err)
		return
	}
	app.jsonResponse(http.StatusOK, w, "OK")
}
