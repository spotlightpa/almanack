package herokuapi

import (
	"context"
	"flag"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/pkg/common"
)

func AddFlags(fl *flag.FlagSet) *Configurator {
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
	vals = make(map[string]string)
	err = requests.URL("https://api.heroku.com").
		Pathf("/apps/%s/config-vars", conf.appName).
		Bearer(conf.apiKey).
		Accept("application/vnd.heroku+json; version=3").
		ToJSON(&vals).
		Fetch(ctx)
	return
}

func listVisitedFlagNames(fl *flag.FlagSet) map[string]bool {
	seen := make(map[string]bool)
	fl.Visit(func(f *flag.Flag) {
		seen[f.Name] = true
	})
	return seen
}

func (conf *Configurator) Configure(f2c map[string]string) error {
	if conf.apiKey == "" {
		common.Logger.Printf("no Heroku API key")
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
		common.Logger.Printf("no missing values for Heroku to enrich")
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
			common.Logger.Printf("%s not set as %s in Heroku", ff.Name, cname)
		} else {
			common.Logger.Printf("setting %s from Heroku", ff.Name)
			errs.Push(conf.fl.Set(ff.Name, val))
		}
	}
	return errs.Merge()
}
