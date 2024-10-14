package db_test

import (
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestJoinUnique(t *testing.T) {
	for _, tc := range []struct{ have, want string }{
		{";", ""},
		{"1;", "1"},
		{";1", "1"},
		{"1;2", "1 2"},
		{"1 2;2 3", "1 2 3"},
		{"1 2 3;2 3 4", "1 2 3 4"},
	} {
		aStr, bStr, _ := strings.Cut(tc.have, ";")
		a, b := strings.Fields(aStr), strings.Fields(bStr)
		got := db.ConcatUnique(a, b, func(s string) string { return s })
		be.AllEqual(t, strings.Fields(tc.want), got)
	}
}
