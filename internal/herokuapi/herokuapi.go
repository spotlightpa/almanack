package herokuapi

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/spotlightpa/almanack/pkg/common"
)

func ConfigureFlagSet(fl *flag.FlagSet) *Configurator {
	conf := Configurator{fl: fl}
	fl.StringVar(&conf.apiKey, "heroku-api-key", "", "`API key` for retrieving config from Heroku")
	fl.StringVar(&conf.appName, "heroku-app-name", "", "`name` for Heroku app to get config from")
	return &conf
}

type Configurator struct {
	fl              *flag.FlagSet
	apiKey, appName string
}

const timeout = 5 * time.Second

func (conf *Configurator) Request() (vals map[string]string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.heroku.com/apps/"+conf.appName+"/config-vars",
		nil,
	)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+conf.apiKey)
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from Heroku: %s", resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	vals = make(map[string]string)
	if err = json.Unmarshal(b, &vals); err != nil {
		return nil, err
	}

	return vals, nil
}

func listVisitedFlagNames(fl *flag.FlagSet) map[string]bool {
	seen := make(map[string]bool)
	fl.Visit(func(f *flag.Flag) {
		seen[f.Name] = true
	})
	return seen
}

func (conf *Configurator) Configure(l common.Logger, f2c map[string]string) error {
	if conf.apiKey == "" {
		l.Printf("no Heroku API key")
		return nil
	}
	seen := listVisitedFlagNames(conf.fl)

	unseen := []*flag.Flag{}
	conf.fl.VisitAll(func(ff *flag.Flag) {
		if !seen[ff.Name] && f2c[ff.Name] != "" {
			unseen = append(unseen, ff)
		}
	})
	if len(unseen) == 0 {
		l.Printf("no missing values for Heroku to enrich")
		return nil
	}
	vals, err := conf.Request()
	if err != nil {
		return err
	}
	var errs errutil.Slice
	for _, ff := range unseen {
		cname := f2c[ff.Name]
		val := vals[cname]
		if val == "" {
			l.Printf("%s not set as %s in Heroku", ff.Name, cname)
		} else {
			l.Printf("setting %s from Heroku", ff.Name)
			errs.Push(conf.fl.Set(ff.Name, val))
		}
	}
	return errs.Merge()
}
