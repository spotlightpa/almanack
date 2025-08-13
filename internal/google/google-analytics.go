package google

import (
	"context"
	"fmt"
	"net/http"

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

	req := RunReportRequest{
		DateRanges: []DateRange{
			{
				StartDate: "today",
				EndDate:   "today",
			},
		},
		Dimensions: []Dimension{
			{Name: "pagePath"},
		},
		Metrics: []Metric{
			{Name: "screenPageViews"},
		},
		OrderBys: []OrderBy{
			{
				Metric: &MetricOrderBy{
					MetricName: "screenPageViews",
				},
				Desc: true,
			},
		},
		DimensionFilter: &FilterClause{
			Filter: &Filter{
				FieldName: "pagePath",
				StringFilter: &StringFilter{
					MatchType: "FULL_REGEXP",
					// TODO: Change this if we have new site sections
					Value: `^/(news|statecollege|berks)/\d\d\d\d/.*`,
				},
			},
		},
		Limit: 20,
	}

	var data RunReportResponse
	if err = requests.
		URL("https://analyticsdata.googleapis.com").
		Pathf("/v1beta/properties/%s:runReport", gsvc.viewID).
		Client(cl).
		BodyJSON(&req).
		ToJSON(&data).
		Fetch(ctx); err != nil {
		return nil, fmt.Errorf("could not get most-popular: %w", err)
	}

	pages = make([]string, 0, len(data.Rows))
	pagesSet := make(map[string]bool, len(data.Rows))
	for _, row := range data.Rows {
		if len(row.DimensionValues) != 1 {
			return nil, fmt.Errorf("got bad row length: %d", len(row.DimensionValues))
		}
		// TODO: We could go through and add up all the query string variants
		// then re-sort, but seems like overkill
		page := row.DimensionValues[0].Value
		if !pagesSet[page] {
			pagesSet[page] = true
			pages = append(pages, page)
		}
	}
	l := almlog.FromContext(ctx)
	l.InfoContext(ctx, "google.MostPopularNews", "count", len(pages))
	return pages, nil
}

type RunReportRequest struct {
	DateRanges      []DateRange   `json:"dateRanges"`
	Dimensions      []Dimension   `json:"dimensions"`
	Metrics         []Metric      `json:"metrics"`
	OrderBys        []OrderBy     `json:"orderBys"`
	DimensionFilter *FilterClause `json:"dimensionFilter,omitempty"`
	Limit           int           `json:"limit"`
}

type DateRange struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type Dimension struct {
	Name string `json:"name"`
}

type Metric struct {
	Name string `json:"name"`
}

type OrderBy struct {
	Metric *MetricOrderBy `json:"metric,omitempty"`
	Desc   bool           `json:"desc"`
}

type MetricOrderBy struct {
	MetricName string `json:"metricName"`
}

type RunReportResponse struct {
	Rows []Row `json:"rows"`
}

type Row struct {
	DimensionValues []DimensionValue `json:"dimensionValues"`
	MetricValues    []MetricValue    `json:"metricValues"`
}

type DimensionValue struct {
	Value string `json:"value"`
}

type MetricValue struct {
	Value string `json:"value"`
}

type FilterClause struct {
	Filter *Filter `json:"filter,omitempty"`
}

type Filter struct {
	FieldName    string        `json:"fieldName"`
	StringFilter *StringFilter `json:"stringFilter,omitempty"`
}

type StringFilter struct {
	MatchType string `json:"matchType"`
	Value     string `json:"value"`
}
