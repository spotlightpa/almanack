package clis

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"

	"github.com/carlmjohnson/flagx"
	"github.com/carlmjohnson/versioninfo"
	"github.com/spotlightpa/nkotb/pkg/gdocs"
	"golang.org/x/net/html"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/docs/v1"
)

const GDocsApp = "gdocs"

func GDocs(args []string) error {
	var app gdocsAppEnv
	err := app.ParseArgs(args)
	if err != nil {
		return err
	}
	err = app.Exec()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func (app *gdocsAppEnv) ParseArgs(args []string) error {
	fl := flag.NewFlagSet(GDocsApp, flag.ContinueOnError)
	fl.StringVar(&app.docid, "id", "", "ID for Google Doc")
	fl.StringVar(&app.outputDoc, "write-doc", "", "`path` to write out document")
	fl.StringVar(&app.inputDoc, "read-doc", "", "`path` to read document from instead of Google Docs")
	fl.StringVar(&app.oauthClientID, "oauth-client-id", "", "client `id` for Google OAuth 2.0 authentication")
	fl.StringVar(&app.oauthClientSecret, "oauth-client-secret", "", "client `secret` for Google OAuth 2.0 authentication")

	app.Logger = log.New(os.Stderr, GDocsApp+" ", log.LstdFlags)
	flagx.BoolFunc(fl, "silent", `don't log debug output`, func() error {
		app.Logger.SetOutput(io.Discard)
		return nil
	})

	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `gdocs %s - extracts a document from Google Docs

Usage:

	gdocs [options]

Uses Google default credentials if no Oauth credentials are provided. See

https://developers.google.com/accounts/docs/application-default-credentials
https://developers.google.com/identity/protocols/oauth2

Options:
`, versioninfo.Version)
		fl.PrintDefaults()
		fmt.Fprintln(fl.Output(), "")
	}
	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagx.ParseEnv(fl, GDocsApp); err != nil {
		return err
	}

	app.docid = gdocs.NormalizeID(app.docid)
	return nil
}

type gdocsAppEnv struct {
	docid             string
	oauthClientID     string
	oauthClientSecret string
	inputDoc          string
	outputDoc         string
	*log.Logger
}

func (app *gdocsAppEnv) Exec() (err error) {
	app.Println("starting Google Docs service")
	ctx := context.Background()
	var doc *docs.Document
	if app.inputDoc == "" {
		getClient := app.oauthClient
		if app.oauthClientID == "" || app.oauthClientSecret == "" {
			getClient = app.defaultCredentials
		}
		client, err := getClient(ctx)
		if err != nil {
			return err
		}
		app.Printf("getting %q", app.docid)
		doc, err = gdocs.Request(ctx, client, app.docid)
		if err != nil {
			return err
		}
		if app.outputDoc != "" {
			b, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				return err
			}

			if err = os.WriteFile(app.outputDoc, b, 0644); err != nil {
				return err
			}
		}
	} else {
		app.Printf("reading %q", app.inputDoc)
		b, err := os.ReadFile(app.inputDoc)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(b, doc); err != nil {
			return err
		}
	}

	app.Printf("got %q", doc.Title)

	n := gdocs.Convert(doc)

	return html.Render(os.Stdout, n)
}

func (app *gdocsAppEnv) defaultCredentials(ctx context.Context) (*http.Client, error) {
	app.Printf("using default credentials")
	return google.DefaultClient(ctx, scopes...)
}

func (app *gdocsAppEnv) oauthClient(ctx context.Context) (client *http.Client, err error) {
	app.Printf("using oauth credentials")
	stateToken, err := makeStateToken()
	if err != nil {
		return nil, err
	}
	code := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("state") != stateToken {
			http.NotFound(w, r)
			return
		}
		code <- r.FormValue("code")
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<h1>Success</h1><p>You may close this window."))
	}))
	defer srv.Close()

	conf := &oauth2.Config{
		ClientID:     app.oauthClientID,
		ClientSecret: app.oauthClientSecret,
		RedirectURL:  srv.URL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
	// Redirect user to Google's consent page to ask for permission
	url := conf.AuthCodeURL(stateToken)
	if launcherr := exec.CommandContext(ctx, "open", url).Run(); launcherr != nil {
		fmt.Printf("Visit the URL for the auth dialog: %v", url)
	}
	tok, err := conf.Exchange(ctx, <-code)
	if err != nil {
		return nil, err
	}
	return conf.Client(ctx, tok), nil
}
