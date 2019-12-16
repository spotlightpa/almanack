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

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/flagext"
	"github.com/gomodule/redigo/redis"
	"github.com/mattbaird/gochimp"
	"github.com/peterbourgon/ff"

	"github.com/spotlightpa/almanack/internal/jsonschema"
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

func parseArgs(args []string) (*app, error) {
	var a app
	rp := redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
	}
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fl.StringVar(&a.srcFeedURL, "src-feed", "", "source URL for Arc feed")
	fl.StringVar(&a.mcapi, "mc-api-key", "", "API `key` for MailChimp")
	fl.StringVar(&a.mclistid, "mc-list-id", "", "List `ID` MailChimp campaign")
	fl.Var(redisflag.Value(&rp.Dial), "redis-url", "`URL` connection string for Redis")
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
	if rp.Dial == nil {
		fmt.Fprint(fl.Output(), "Must set -redis-url\n\n")
		fl.Usage()
		return nil, flag.ErrHelp
	}
	a.rp = &rp
	return &a, nil
}

type app struct {
	srcFeedURL string
	mcapi      string
	mclistid   string
	rp         *redis.Pool
	*log.Logger
}

func (a *app) exec() error {
	a.Println("starting", AppName)
	start := time.Now()
	defer func() { a.Println("finished in", time.Since(start)) }()

	a.Println("fetching", a.srcFeedURL)
	var newfeed, oldfeed jsonschema.API
	if err := a.fetchJSON(a.srcFeedURL, &newfeed); err != nil {
		return err
	}

	a.Println("checking redis")
	err := a.GetSet("almanack-worker.feed", &oldfeed, &newfeed)
	if err == redis.ErrNil {
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

func (a *app) fetchJSON(url string, v interface{}) error {
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

// Ping Redis
func (a *app) Ping() (err error) {
	a.Println("Ping Redis")
	conn := a.rp.Get()
	defer errutil.Defer(&err, conn.Close)

	_, err = conn.Do("PING")
	return
}

// GetSet converts values to JSON bytes and calls GETSET in Redis
func (a *app) GetSet(key string, getv, setv interface{}) (err error) {
	a.Printf("Redis GETSET %q", key)
	conn := a.rp.Get()
	defer errutil.Defer(&err, conn.Close)

	setb, err := json.Marshal(setv)
	if err != nil {
		return err
	}
	getb, err := redis.Bytes(conn.Do("GETSET", key, setb))
	if err != nil {
		return err
	}
	return json.Unmarshal(getb, getv)
}

func diffFeed(newfeed, oldfeed jsonschema.API) []jsonschema.Contents {
	readyids := make(map[string]bool, len(oldfeed.Contents))
	for _, story := range oldfeed.Contents {
		if story.Workflow.StatusCode >= jsonschema.StatusReady {
			readyids[story.ID] = true
		}
	}
	var newstories []jsonschema.Contents
	for _, story := range newfeed.Contents {
		if story.Workflow.StatusCode >= jsonschema.StatusReady &&
			!readyids[story.ID] {
			newstories = append(newstories, story)
		}
	}
	return newstories
}

func (a *app) SendCampaign(subject, body string) error {
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

https://spotlightpa-almanack.netlify.com/articles/{{ .ID }}

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

func (a *app) makeMessage(diff []jsonschema.Contents) (subject, body string) {
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
