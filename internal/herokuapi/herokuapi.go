package herokuapi

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/spotlightpa/almanack/pkg/common"
)

func FromFlagSet(fl *flag.FlagSet) *Config {
	var conf Config
	fl.StringVar(&conf.apiKey, "heroku-api-key", "", "`API key` for retrieving config from Heroku")
	fl.StringVar(&conf.appName, "heroku-app-name", "", "`name` for Heroku app to get config from")
	return &conf
}

type Config struct {
	apiKey, appName string
}

const timeout = 5 * time.Second

func (conf *Config) Request() (vals map[string]string, err error) {
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
	b, err := ioutil.ReadAll(resp.Body)
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

func (conf *Config) GetConfig(fl *flag.FlagSet, l common.Logger, f2c map[string]string) error {
	if conf.apiKey == "" {
		l.Printf("no Heroku API key")
		return nil
	}
	seen := listVisitedFlagNames(fl)

	unseen := []*flag.Flag{}
	fl.VisitAll(func(ff *flag.Flag) {
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
			errs.Push(fl.Set(ff.Name, val))
		}
	}
	return errs.Merge()
}
