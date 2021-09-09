package timeutil_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/spotlightpa/almanack/internal/timeutil"
)

func TestEqualish(t *testing.T) {
	parseTime := func(s string) sql.NullTime {
		t, err := time.Parse("15:04:05", s)
		return sql.NullTime{Time: t, Valid: err == nil}
	}
	cases := map[string]struct {
		a, b string
		want bool
	}{
		"both null":          {"", "", true},
		"first null":         {"", "1:00:00", false},
		"second null":        {"1:00:00", "", false},
		"same":               {"1:00:00", "1:00:00", true},
		"first minute later": {"1:01:00", "1:00:00", true},
		"second later":       {"1:00:00", "1:01:00", true},
		"first much later":   {"1:10:00", "1:00:00", false},
		"second much later":  {"1:00:00", "1:10:00", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			a, b := parseTime(tc.a), parseTime(tc.b)
			got := timeutil.Equalish(a, b)
			if got != tc.want {
				t.Errorf("timeutil.Equalish(%q, %q) != %v", tc.a, tc.b, tc.want)
			}
		})
	}
}
