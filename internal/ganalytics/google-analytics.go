package ganalytics

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/carlmjohnson/flagext"
	"github.com/spotlightpa/almanack/internal/httpjson"
	"github.com/spotlightpa/almanack/pkg/common"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func Var(fl *flag.FlagSet) func(l common.Logger) *Client {
	var ga Client
	// Using a crazy Base64+GZIP because storing JSON containing \n in
	//an env var breaks a lot for some reason
	flagext.Callback(fl, "google-json", "", "GZIP Base64 JSON credentials for Google",
		func(s string) error {
			b, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return err
			}
			g, err := gzip.NewReader(bytes.NewReader(b))
			if err != nil {
				return err
			}
			defer g.Close()
			b, err = io.ReadAll(g)
			if err != nil {
				return err
			}
			creds, err := google.CredentialsFromJSON(
				oauth2.NoContext, b, "https://www.googleapis.com/auth/analytics.readonly",
			)
			if err != nil {
				return err
			}
			ga.c = oauth2.NewClient(oauth2.NoContext, creds.TokenSource)

			return nil
		})
	fl.StringVar(&ga.viewID, "ga-view-id", "", "view `ID` for Google Analytics")

	return func(c common.Logger) *Client {
		ga.l = c
		return &ga
	}
}

type Client struct {
	c      *http.Client
	l      common.Logger
	viewID string
}

func (ga *Client) getClient(ctx context.Context) (err error) {
	if ga.viewID == "" {
		return fmt.Errorf("view ID not set")
	}
	if ga.c != nil {
		return nil
	}
	ga.l.Printf("falling back to default Google credentials")
	ga.c, err = google.DefaultClient(ctx, "https://www.googleapis.com/auth/analytics.readonly")
	return
}

func (ga *Client) MostPopularNews(ctx context.Context) ([]string, error) {
	if err := ga.getClient(ctx); err != nil {
		return nil, err
	}

	req := &AnalyticsRequest{
		ReportRequests: []ReportRequest{{
			ViewID: ga.viewID,
			Metrics: []Metric{{
				Expression: "ga:uniquePageviews",
			}},
			Dimensions: []Dimension{{
				Name: "ga:pagePath",
			}},
			DateRanges: []DateRange{{
				StartDate: "today",
				EndDate:   "today",
			}},
			OrderBys: []OrderBy{{
				FieldName: "ga:uniquePageviews",
				SortOrder: "DESCENDING",
			}},
			FiltersExpression: `ga:pagePath=~^/news/\d\d\d\d`,
			PageSize:          20,
		}},
	}

	var data AnalyticsResponse
	if err := httpjson.Post(
		ctx,
		ga.c,
		"https://analyticsreporting.googleapis.com/v4/reports:batchGet",
		req,
		&data,
	); err != nil {
		return nil, fmt.Errorf("could not get most-popular: %w", err)
	}

	if len(data.Reports) != 1 {
		return nil, fmt.Errorf("got bad report length: %d", len(data.Reports))
	}
	report := &data.Reports[0]
	pages := make([]string, 0, len(report.Data.Rows))
	pagesSet := make(map[string]bool, len(report.Data.Rows))
	for _, row := range report.Data.Rows {
		if len(row.Dimensions) != 1 {
			return nil, fmt.Errorf("got bad row length: %d", len(row.Dimensions))
		}
		page := row.Dimensions[0]
		u, err := url.Parse(page)
		if err != nil {
			continue
		}
		// TODO: We could go through and add up all the query string variants
		// then re-sort, but seems like overkill
		page = u.Path
		if !pagesSet[page] {
			pagesSet[page] = true
			pages = append(pages, page)
		}
	}
	ga.l.Printf("got %d most-popular pages", len(pages))
	return pages, nil
}

type AnalyticsRequest struct {
	ReportRequests []ReportRequest `json:"reportRequests"`
}

type ReportRequest struct {
	ViewID            string      `json:"viewId"`
	DateRanges        []DateRange `json:"dateRanges"`
	Dimensions        []Dimension `json:"dimensions"`
	Metrics           []Metric    `json:"metrics"`
	FiltersExpression string      `json:"filtersExpression"`
	OrderBys          []OrderBy   `json:"orderBys"`
	PageSize          int         `json:"pageSize"`
	PageToken         string      `json:"pageToken"`
}

type DateRange struct {
	EndDate   string `json:"endDate"`
	StartDate string `json:"startDate"`
}

type Dimension struct {
	Name string `json:"name"`
}

type Metric struct {
	Expression string `json:"expression"`
}

type OrderBy struct {
	FieldName string `json:"fieldName"`
	SortOrder string `json:"sortOrder"`
}

type AnalyticsResponse struct {
	Reports []Report `json:"reports"`
}

type Report struct {
	ColumnHeader ColumnHeader `json:"columnHeader"`
	Data         Data         `json:"data"`
}

type Data struct {
	Rows     []Row    `json:"rows"`
	Totals   []Values `json:"totals"`
	RowCount int      `json:"rowCount"`
	Minimums []Values `json:"minimums"`
	Maximums []Values `json:"maximums"`
}

type MetricHeaderEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type MetricHeader struct {
	MetricHeaderEntries []MetricHeaderEntry `json:"metricHeaderEntries"`
}

type ColumnHeader struct {
	Dimensions   []string     `json:"dimensions"`
	MetricHeader MetricHeader `json:"metricHeader"`
}

type Values struct {
	Values []string `json:"values"`
}

type Row struct {
	Dimensions []string `json:"dimensions"`
	Metrics    []Values `json:"metrics"`
}
