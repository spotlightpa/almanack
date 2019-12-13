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
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	fl.StringVar(&a.srcFeedURL, "src-feed", "", "source URL for Arc feed")
	fl.StringVar(&a.mcapi, "mc-api-key", "", "API key for MailChimp")
	fl.StringVar(&a.mclistid, "mc-list-id", "", "List ID MailChimp campaign")
	redaddr := fl.String("redis-address", "", `Address for Redis connection pool`)
	redpassword := fl.String("redis-password", "", `Password for Redis connection pool`)
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

	a.rp = newPool(*redaddr, *redpassword)

	return &a, nil
}

type app struct {
	srcFeedURL string
	mcapi      string
	mclistid   string
	rp         *redis.Pool
	*log.Logger
}

func newPool(addr, password string) *redis.Pool {
	if addr == "" {
		return nil
	}
	dialer := func() (redis.Conn, error) { return redis.Dial("tcp", addr) }
	if password != "" {
		dialer = func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		}
	}
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        dialer,
	}
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

	// newfeed.Contents[2].Workflow.StatusCode = 0

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

// GetSet converts values to JSON bytes and calls GETSET in Redis
func (a *app) GetSet(key string, getv, setv interface{}) (err error) {
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
