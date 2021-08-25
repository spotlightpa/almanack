package stringutils

import (
	"regexp"
	"strings"
)

// Return first non-blank string
func First(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// Cut cuts s around the first instance of sep,
// returning the text before and after sep.
// The found result reports whether sep appears in s.
// If sep does not appear in s, cut returns s, "", false.
//
// See https://github.com/golang/go/issues/46336
func Cut(s, sep string) (before, after string, found bool) {
	if i := strings.Index(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return s, "", false
}

var (
	articleRe   = regexp.MustCompile(`\b(the|an?)\b`)
	pennRe      = regexp.MustCompile(`\bpa\b`)
	possesiveRe = regexp.MustCompile(`\.?[’']s`)
	nonasciiRe  = regexp.MustCompile(`\W+`)
)

func Slugify(s string) string {
	s = strings.ToLower(s)
	s = articleRe.ReplaceAllString(s, " ")
	s = pennRe.ReplaceAllString(s, "pennsylvania")
	s = possesiveRe.ReplaceAllString(s, "s")
	s = nonasciiRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = nonasciiRe.ReplaceAllString(s, "-")
	return s
}
