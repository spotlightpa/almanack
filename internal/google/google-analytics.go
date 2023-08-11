package google

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/requests"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

func (gsvc *Service) GAClient(ctx context.Context) (cl *http.Client, err error) {
	return gsvc.client(ctx, "https://www.googleapis.com/auth/analytics.readonly")
}

func (gsvc *Service) MostPopularNews(ctx context.Context, cl *http.Client) (pages []string, err error) {
	defer errorx.Trace(&err)

	if gsvc.viewID == "" {
		return nil, fmt.Errorf("view ID not set")
	}

	req := &AnalyticsRequest{
		ReportRequests: []ReportRequest{{
			ViewID: gsvc.viewID,
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
			FiltersExpression: `ga:pagePath=~^/(news|statecollege)/\d\d\d\d`,
			PageSize:          20,
		}},
	}

	var data AnalyticsResponse
	if err = requests.
		URL("https://analyticsreporting.googleapis.com/v4/reports:batchGet").
		Client(cl).
		BodyJSON(req).
		ToJSON(&data).
		Fetch(ctx); err != nil {
		return nil, fmt.Errorf("could not get most-popular: %w", err)
	}

	if len(data.Reports) != 1 {
		return nil, fmt.Errorf("got bad report length: %d", len(data.Reports))
	}
	report := &data.Reports[0]
	pages = make([]string, 0, len(report.Data.Rows))
	pagesSet := make(map[string]bool, len(report.Data.Rows))
	for _, row := range report.Data.Rows {
		if len(row.Dimensions) != 1 {
			return nil, fmt.Errorf("got bad row length: %d", len(row.Dimensions))
		}
		page := row.Dimensions[0]
		u, err2 := url.Parse(page)
		if err2 != nil {
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
	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "google.MostPopularNews", "count", len(pages))
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
