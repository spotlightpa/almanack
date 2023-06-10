// Package stringx has string utilities
package stringx

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/spotlightpa/almanack/internal/syncx"
)

// First non-blank string
func First(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func mustBeLazy(s string) func() *regexp.Regexp {
	return syncx.Once(
		func() *regexp.Regexp {
			return regexp.MustCompile(s)
		})
}

var (
	articleRe   = mustBeLazy(`\b(the|an?)\b`)
	pennRe      = mustBeLazy(`\bpa\b`)
	possesiveRe = mustBeLazy(`\.?[’']s`)
	nonasciiRe  = mustBeLazy(`\W+`)
)

func SlugifyURL(s string) string {
	s = strings.ToLower(s)
	s = articleRe().ReplaceAllString(s, " ")
	s = pennRe().ReplaceAllString(s, "pennsylvania")
	s = possesiveRe().ReplaceAllString(s, "s")
	s = nonasciiRe().ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = nonasciiRe().ReplaceAllString(s, "-")
	return s
}

func SlugifyFilename(s string) string {
	hadDash := false
	f := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			hadDash = false
			return r - 'A' + 'a'
		case
			r >= 'a' && r <= 'z',
			r >= '0' && r <= '9',
			r == '.':
			hadDash = false
			return r
		case hadDash:
			return -1
		}
		hadDash = true
		return '-'
	}
	return strings.Map(f, s)
}

func LastCut(s, sep string) (before, after string, found bool) {
	if i := strings.LastIndex(s, sep); i >= 0 {
		return s[:i], s[i+len(sep):], true
	}
	return "", s, false
}

// RemoveParens removes parentheses and any text that is inside of them.
func RemoveParens(s string) string {
	if !strings.Contains(s, "(") {
		return s
	}

	var sb strings.Builder
	sb.Grow(len(s))
	openCount := 0

	for _, ch := range []byte(s) {
		if ch == '(' {
			openCount++
		} else if ch == ')' {
			if openCount > 0 {
				openCount--
			}
		} else if openCount == 0 {
			sb.WriteByte(ch)
		}
	}

	return sb.String()
}

func RemoveAllWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}
