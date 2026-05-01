package almapp

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/earthboundkid/resperr/v2"
	"github.com/earthboundkid/slackhook/v2"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/jwthook"
	"github.com/spotlightpa/almanack/internal/services/netlifyid"
)

func (app *appEnv) notFound(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	app.replyErr(w, r, resperr.NotFound(r))
}

func (app *appEnv) ping(w http.ResponseWriter, r *http.Request) {
	app.logStart(r)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "no-store")
	b, err := httputil.DumpRequest(r, true)
	if err != nil {
		app.replyErr(w, r, err)
		return
	}

	w.Write(b)
}

func (app *appEnv) pingErr(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	statusCode, _ := strconv.Atoi(code)
	app.logStart(r, "code", code)

	app.replyNewErr(statusCode, w, r, "got test ping %q", code)
}

func (app *appEnv) getBookmarklet(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	app.logStart(r, "slug", slug)

	page, err := app.svc.Queries.GetPageByURLPath(r.Context(), "%"+slug)
	if err != nil {
		if db.IsNotFound(err) {
			err = resperr.E{
				E: db.NoRowsAs404(err, "could not find url-path %q", slug),
				M: fmt.Sprintf("No matching page found for %s.", slug)}
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
	l := almlog.FromContext(r.Context())
	app.svc.SlackTech.Post(r.Context(), l.InfoContext, app.svc.Client,
		slackhook.Message{
			Attachments: []slackhook.Attachment{
				{
					Title: "New Almanack Registration",
					Text:  msg,
					Color: color,
					Fields: []slackhook.Field{
						{
							Title: "Roles",
							Value: strings.Join(req.User.AppMetadata.Roles, ", "),
							Short: true,
						}}}}},
	)
	app.replyJSON(http.StatusOK, w, req.User)
}
