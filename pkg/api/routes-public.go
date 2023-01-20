package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/carlmjohnson/requests"
	"github.com/carlmjohnson/resperr"
	"github.com/go-chi/chi/v5"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/httpx"
	"github.com/spotlightpa/almanack/internal/jwthook"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"golang.org/x/exp/slog"
)

func (app *appEnv) notFound(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	app.replyErr(w, r, resperr.NotFound(r))
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=60")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	w.Write(b)
}

func (app *appEnv) pingErr(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	statusCode, _ := strconv.Atoi(code)
	app.logStart(r, "code", code)

	app.replyNewErr(statusCode, w, r, "got test ping %q", code)
}

var inkyURL = must.Get(url.Parse("https://www.inquirer.com"))

func (app *appEnv) getProxyImage(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	encURL := chi.URLParam(r, "encURL")
	decURL, err := base64.URLEncoding.DecodeString(encURL)
	if err != nil {
		app.replyErr(w, r, resperr.WithUserMessagef(
			nil, "Could not decode URL param: %q", encURL,
		))
		return
	}
	u, err := inkyURL.Parse(string(decURL))
	if err != nil {
		app.replyErr(w, r, resperr.WithUserMessagef(
			nil, "Bad image URL: %q", decURL,
		))
		return
	}

	srcURL := u.String()
	l := slog.FromContext(r.Context())

	dbImage, err := app.svc.Queries.GetImageBySourceURL(r.Context(), srcURL)
	switch {
	case db.IsNotFound(err):
		l.Info("getProxyImage: image not found", "src", srcURL)

	case err != nil:
		app.replyHTMLErr(w, r, err)
		return

	case err == nil && !dbImage.IsUploaded:
		l.Info("getProxyImage: image found but awaiting upload", "src", srcURL)

	case err == nil && dbImage.IsUploaded:
		l.Info("getProxyImage: redirecting", "src", srcURL)
		redirect, err := app.svc.ImageStore.SignGetURL(r.Context(), dbImage.Path)
		if err != nil {
			app.logErr(r.Context(), err)
			app.replyHTMLErr(w, r, err)
			return
		}
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}

	l.Info("getProxyImage: proxying", "src", srcURL)

	const urlWhitelist = `^https://[^/]*(.inquirer.com|.arcpublishing.com|arc-anglerfish-arc2-prod-pmn.s3.amazonaws.com)/`
	cl := *app.svc.Client
	cl.Transport = requests.PermitURLTransport(cl.Transport, urlWhitelist)
	body, ctype, err := almanack.FetchImageURL(r.Context(), &cl, u.String())
	if err != nil {
		app.replyHTMLErr(w, r, err)
		return
	}
	w.Header().Set("Content-Type", ctype)
	ext := strings.TrimPrefix(ctype, "image/")
	httpx.SetAttachmentName(w.Header(), "image."+ext)
	w.Header().Set("Cache-Control", "public, max-age=900")
	if _, err = w.Write(body); err != nil {
		app.logErr(r.Context(), err)
	}
}

func (app *appEnv) getBookmarklet(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	app.logStart(r, "slug", slug)

	page, err := app.svc.Queries.GetPageByURLPath(r.Context(), "%"+slug+"%")
	if err != nil {
		if db.IsNotFound(err) {
			err = resperr.WithUserMessagef(
				db.NoRowsAs404(err, "could not find url-path %q", slug),
				"No matching page found for %s.", slug)
		}
		app.replyHTMLErr(w, r, err)
		return
	}
	http.Redirect(w, r,
		fmt.Sprintf("/admin/news/%d", page.ID),
		http.StatusTemporaryRedirect)
}

func (app *appEnv) postIdentityHook(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)

	var req struct {
		EventType string         `json:"event"`
		User      netlifyid.User `json:"user"`
	}
	err := jwthook.VerifyRequest(r,
		app.svc.NetlifyWebhookSecret, "d4cce6f2-6b46-4bba-b126-cfb8f469e3c5", "gotrue",
		time.Now(),
		&req)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	if req.EventType != "signup" {
		app.replyNewErr(http.StatusBadRequest, w, r,
			"unexpect event type: %q", req.EventType)
		return
	}
	roles, err := db.GetRolesForEmail(r.Context(), app.svc.Queries, req.User.Email)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}
	req.User.AppMetadata.Roles = append(req.User.AppMetadata.Roles, roles...)

	const (
		colorGreen = "#78bc20"
		colorRed   = "#da291c"
	)

	msg := fmt.Sprintf("%s <%s> with %d role(s)",
		req.User.UserMetadata.FullName,
		req.User.Email,
		len(req.User.AppMetadata.Roles))
	color := colorGreen
	if len(req.User.AppMetadata.Roles) < 1 {
		color = colorRed
	}
	app.svc.SlackTech.Post(context.Background(),
		slack.Message{
			Attachments: []slack.Attachment{
				{
					Title: "New Almanack Registration",
					Text:  msg,
					Color: color,
					Fields: []slack.Field{
						{
							Title: "Roles",
							Value: strings.Join(req.User.AppMetadata.Roles, ", "),
							Short: true,
						}}}}},
	)
	app.replyJSON(http.StatusOK, w, req.User)
}
