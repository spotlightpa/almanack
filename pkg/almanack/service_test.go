package almanack

import (
	"database/sql"
	"testing"
	"time"
)

func TestDiffTime(t *testing.T) {
	parseTime := func(s string) sql.NullTime {
		t, err := time.Parse("15:04:05", s)
		return sql.NullTime{Time: t, Valid: err == nil}
	}
	cases := map[string]struct {
		a, b string
		want bool
	}{
		"both null":          {"", "", false},
		"first null":         {"", "1:00:00", true},
		"second null":        {"1:00:00", "", true},
		"same":               {"1:00:00", "1:00:00", false},
		"first minute later": {"1:01:00", "1:00:00", false},
		"second later":       {"1:00:00", "1:01:00", false},
		"first much later":   {"1:10:00", "1:00:00", true},
		"second much later":  {"1:00:00", "1:10:00", true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			a, b := parseTime(tc.a), parseTime(tc.b)
			got := diffTime(a, b)
			if got != tc.want {
				t.Errorf("timeDiff(%q, %q) != %v", tc.a, tc.b, tc.want)
			}
		})
	}
}
