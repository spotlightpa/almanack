package worker

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/carlmjohnson/flagext"
	"github.com/mattbaird/gochimp"
	"github.com/peterbourgon/ff"

	"github.com/spotlightpa/almanack/internal/errutil"
	"github.com/spotlightpa/almanack/internal/filestore"
	"github.com/spotlightpa/almanack/internal/jsonschema"
	"github.com/spotlightpa/almanack/internal/redis"
	"github.com/spotlightpa/almanack/internal/redisflag"
)

const AppName = "almanack-worker"

func CLI(args []string) error {
	a, err := parseArgs(args)
	if err != nil {
		return err
	}
	if err := a.exec(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return err
	}
	return nil
}

func parseArgs(args []string) (*appEnv, error) {
	var a appEnv
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fl.StringVar(&a.srcFeedURL, "src-feed", "", "source URL for Arc feed")
	fl.StringVar(&a.mcapi, "mc-api-key", "", "API `key` for MailChimp")
	fl.StringVar(&a.mclistid, "mc-list-id", "", "List `ID` MailChimp campaign")
	getDialer := redisflag.Var(fl, "redis-url", "`URL` connection string for Redis")
	a.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	fl.Var(
		flagext.Logger(a.Logger, flagext.LogSilent),
		"silent",
		`don't log debug output`,
	)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `almanack-worker help

Options:
`)
		fl.PrintDefaults()
	}
	if err := ff.Parse(fl, args, ff.WithEnvVarPrefix("ALMANACK")); err != nil {
		return nil, err
	}
	if d := getDialer(); d != nil {
		var err error
		if a.store, err = redis.New(d, a.Logger); err != nil {
			return nil, err
		}
	} else {
		a.store = filestore.New("", AppName, a.Logger)
	}
	return &a, nil
}

type appEnv struct {
	srcFeedURL string
	mcapi      string
	mclistid   string
	store      getsetter
	*log.Logger
}

type getsetter interface {
	GetSet(key string, getv, setv interface{}) (err error)
}

func (a *appEnv) exec() error {
	a.Println("starting", AppName)
	start := time.Now()
	defer func() { a.Println("finished in", time.Since(start)) }()

	a.Println("fetching", a.srcFeedURL)
	var newfeed, oldfeed jsonschema.API
	if err := a.fetchJSON(a.srcFeedURL, &newfeed); err != nil {
		return err
	}

	a.Println("checking redis")
	err := a.store.GetSet("almanack-worker.feed", &oldfeed, &newfeed)
	if errutil.Is(err, errutil.NotFound) {
		a.Println("cache miss for old feed")
		return nil
	}
	if err != nil {
		return err
	}

	newstories := diffFeed(newfeed, oldfeed)
	a.Printf("got %d newly ready stories", len(newstories))
	if len(newstories) > 0 {
		subject, body := a.makeMessage(newstories)
		a.Printf("sending %q", subject)
		return a.SendCampaign(subject, body)
	}

	return nil
}

func (a *appEnv) fetchJSON(url string, v interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, v); err != nil {
		return err
	}

	return nil
}

func diffFeed(newfeed, oldfeed jsonschema.API) []jsonschema.Contents {
	readyids := make(map[string]bool, len(oldfeed.Contents))
	for _, story := range oldfeed.Contents {
		if story.Workflow.StatusCode >= jsonschema.StatusSlot {
			readyids[story.ID] = true
		}
	}
	var newstories []jsonschema.Contents
	for _, story := range newfeed.Contents {
		if story.Workflow.StatusCode >= jsonschema.StatusSlot &&
			!readyids[story.ID] {
			newstories = append(newstories, story)
		}
	}
	return newstories
}

func (a *appEnv) SendCampaign(subject, body string) error {
	// Using MC APIv2 because in v3 they decided REST means
	// not being able to create and send a campign in any efficient way
	chimp := gochimp.NewChimp(a.mcapi, true)
	resp, err := chimp.CampaignCreate(gochimp.CampaignCreate{
		Type: "plaintext",
		Options: gochimp.CampaignCreateOptions{
			Subject:   subject,
			ListID:    a.mclistid,
			FromEmail: "press@spotlightpa.org",
			FromName:  "Spotlight PA",
		},
		Content: gochimp.CampaignCreateContent{
			Text: body,
		},
	})
	if err != nil {
		return err
	}
	a.Printf("created campaign %q", resp.Id)
	resp2, err := chimp.CampaignSend(resp.Id)
	a.Printf("sent %v", resp2.Complete)
	return err
}

const messageTemplate = `
{{- range . -}}
{{ .Slug }} now available

https://almanack.data.spotlightpa.org/articles/{{ .ID }}

Planned for {{ .Planning.Scheduling.PlannedPublishDate.Format "January 2, 2006" }}

{{ with .Planning.InternalNote -}}
Publication Notes:

{{ . }}
{{ end -}}

Budget:

{{ .Planning.BudgetLine }}

Word count planned: {{ .Planning.StoryLength.WordCountPlanned}}
Word count actual: {{ .Planning.StoryLength.WordCountActual}}
Lines: {{ .Planning.StoryLength.LineCountActual}}
Column inches: {{ .Planning.StoryLength.InchCountActual}}


{{ end -}}
`

func (a *appEnv) makeMessage(diff []jsonschema.Contents) (subject, body string) {
	slugs := make([]string, len(diff))
	for i := range diff {
		slugs[i] = diff[i].Slug
	}
	subject = fmt.Sprintf(
		"%s now available on Spotlight PA Almanack",
		strings.Join(slugs, ", "))
	t := template.Must(template.New("").Parse(messageTemplate))
	var buf strings.Builder
	if err := t.Execute(&buf, diff); err != nil {
		panic(err)
	}
	body = buf.String()
	return
}
