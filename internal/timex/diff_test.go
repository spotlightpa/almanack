package timex_test

import (
	"testing"
	"time"

	"github.com/carlmjohnson/be"
	"github.com/jackc/pgtype"
	"github.com/spotlightpa/almanack/internal/timex"
)

func TestEqualish(t *testing.T) {
	parseTime := func(s string) pgtype.Timestamptz {
		if s == "0" {
			return pgtype.Timestamptz{Status: pgtype.Present}
		}
		t, err := time.Parse("15:04:05", s)
		status := pgtype.Present
		if err != nil {
			status = pgtype.Null
		}
		return pgtype.Timestamptz{Time: t, Status: status}
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
		"first zero":         {"0", "1:00:00", false},
		"second zero":        {"1:00:00", "0", false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			a, b := parseTime(tc.a), parseTime(tc.b)
			got := timex.Equalish(a, b)
			be.Equal(t, tc.want, got)
		})
	}
}
