// Package plausible gets analytics from Plausible.io.
package plausible

import (
	"context"
	"flag"
	"net/http"
	"regexp"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/internal/syncx"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

type API struct {
	SiteID, Token string
}

func (api *API) AddFlags(fl *flag.FlagSet) {
	fl.StringVar(&api.SiteID, "plausible-site-id", "", "`site ID` for Plausible.io")
	fl.StringVar(&api.Token, "plausible-token", "", "API `token` for Plausible.io")
}

var lazyRe = syncx.Once(func() *regexp.Regexp {
	return regexp.MustCompile(`/(news|statecollege)/\d{4}/\d\d/[\w-]+/`)
})

func (api *API) MostPopularNews(ctx context.Context, cl *http.Client) (pages []string, err error) {
	defer errorx.Trace(&err)

	var res Response
	err = requests.
		URL("https://plausible.io/api/v1/stats/breakdown?period=day&property=event:page&limit=20").
		Param("site_id", api.SiteID).
		Bearer(api.Token).
		Client(cl).
		ToJSON(&res).
		Fetch(ctx)
	if err != nil {
		return nil, err
	}

	re := lazyRe()
	for _, result := range res.Results {
		if re.MatchString(result.Page) {
			pages = append(pages, result.Page)
		}
	}
	l := almlog.FromContext(ctx)
	l.InfoCtx(ctx, "plausible.MostPopularNews", "count", len(pages))
	return pages, nil
}

type Response struct {
	Results []Result `json:"results"`
}

type Result struct {
	Page     string `json:"page"`
	Visitors int    `json:"visitors"`
}