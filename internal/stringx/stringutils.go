package stringx

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
