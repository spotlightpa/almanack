package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/httpx"
	"github.com/spotlightpa/almanack/internal/jwthook"
	"github.com/spotlightpa/almanack/internal/must"
	"github.com/spotlightpa/almanack/internal/netlifyid"
	"github.com/spotlightpa/almanack/internal/slack"
	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/almlog"
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
	code := httpx.PathValue(r, "code")
	statusCode, _ := strconv.Atoi(code)
	app.logStart(r, "code", code)

	app.replyNewErr(statusCode, w, r, "got test ping %q", code)
}

var inkyURL = must.Get(url.Parse("https://www.inquirer.com"))

var imageWhitelist = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`^https://[^/]*(\.inquirer\.com|\.arcpublishing\.com|arc-anglerfish-arc2-prod-pmn\.s3\.amazonaws\.com)/`)
})

func (app *appEnv) getArcImage(w http.ResponseWriter, r *http.Request) http.Handler {
	srcURL := r.URL.Query().Get("src_url")
	app.logStart(r, "src_url", srcURL)

	u, err := inkyURL.Parse(srcURL)
	if err != nil {
		return app.jsonBadRequest(err, "Bad image URL: %q", srcURL)
	}

	srcURL = u.String()

	if !imageWhitelist().MatchString(srcURL) {
		err = fmt.Errorf("srcURL not in imageWhitelist")
		return app.jsonBadRequest(err, "Bad image URL: %q", srcURL)
	}

	l := almlog.FromContext(r.Context())

	dbImage, err := app.svc.Queries.GetImageBySourceURL(r.Context(), srcURL)
	switch {
	case db.IsNotFound(err):
		l.InfoContext(r.Context(), "getProxyImage: image not found", "src", srcURL)

	case err != nil:
		return app.htmlErr(err)

	case err == nil && !dbImage.IsUploaded:
		l.InfoContext(r.Context(), "getProxyImage: image found but awaiting upload", "src", srcURL)

	case err == nil && dbImage.IsUploaded:
		l.InfoContext(r.Context(), "getProxyImage: redirecting", "src", srcURL)
		redirect, err := app.svc.ImageStore.SignGetURL(r.Context(), dbImage.Path)
		if err != nil {
			app.logErr(r.Context(), err)
			return app.htmlErr(err)
		}
		http.Redirect(w, r, redirect, http.StatusFound)
		return nil
	}

	l.InfoContext(r.Context(), "getProxyImage: proxying", "src", srcURL)

	body, ctype, err := almanack.FetchImageURL(r.Context(), app.svc.Client, u.String())
	if err != nil {
		return app.htmlErr(err)
	}
	w.Header().Set("Content-Type", ctype)
	ext := strings.TrimPrefix(ctype, "image/")
	httpx.SetAttachmentName(w.Header(), "image."+ext)
	w.Header().Set("Cache-Control", "public, max-age=900")
	if _, err = w.Write(body); err != nil {
		app.logErr(r.Context(), err)
	}
	return nil
}

func (app *appEnv) getBookmarklet(w http.ResponseWriter, r *http.Request) {
	slug := httpx.PathValue(r, "slug")
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
