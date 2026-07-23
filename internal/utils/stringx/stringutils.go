// Package stringx has string utilities
package stringx

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/spotlightpa/almanack/internal/utils/lazy"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	articleRe   = lazy.RE(`\b(the|an?)\b`)
	pennRe      = lazy.RE(`\bpa\b`)
	possesiveRe = lazy.RE(`\.?[’']s`)
	nonasciiRe  = lazy.RE(`\W+`)
)

func StripAccents(text string) (string, error) {
	t := transform.Chain(
		norm.NFKD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC)
	result, _, err := transform.String(t, text)
	if err != nil {
		return text, err
	}
	result = strings.NewReplacer(
		"æ", "ae",
		"Æ", "AE",
		"œ", "oe",
		"Œ", "OE",
		"ß", "ss",
	).Replace(result)
	return result, nil
}

func SlugifyURL(s string) string {
	s, _ = StripAccents(s)
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
	s, _ = StripAccents(s)
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

func ReplaceWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return '_'
		}
		return r
	}, s)
}

var (
	staffRe        = lazy.RE(`(?i)\bstaff\b`)
	extractSplitRe = lazy.RE(`(?i)[,;]|\b(and|y)\b`)
	outletRe       = lazy.RE(`(?i)\b(of|for|de)\b.*$`)
)

func ExtractNames(s string) []string {
	if result, _, err := transform.String(transform.Chain(
		norm.NFKD,
		runes.Remove(runes.In(unicode.Cf)),
		norm.NFC,
	), s); err == nil {
		s = result
	}

	nameParts := extractSplitRe().Split(s, -1)

	var names []string
	for _, part := range nameParts {
		part = strings.TrimSpace(part)
		if part == "" || staffRe().MatchString(part) {
			continue
		}
		part, _, _ = strings.Cut(part, "/")
		part = outletRe().ReplaceAllString(part, "")
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		names = append(names, part)
	}

	return names
}

func Truncate(s string, maximum int) string {
	r := []rune(s)
	if len(r) <= maximum {
		return s
	}
	truncated := r[:max(maximum-3, 0)]
	return string(truncated) + "…"
}

// FlattenMap converts a map[string]string to a []string of alternating
// key-value pairs sorted by key.
func FlattenMap(m map[string]string) []string {
	result := make([]string, 0, len(m)*2)
	for _, k := range slices.Sorted(maps.Keys(m)) {
		result = append(result, k, m[k])
	}
	return result
}

// FlattenMultimap converts a map[string][]string to a []string of alternating
// key-value pairs sorted by key, repeating the key for each value.
func FlattenMultimap(m map[string][]string) []string {
	size := 0
	for _, vs := range m {
		size += len(vs)
	}
	result := make([]string, 0, size*2)
	for _, k := range slices.Sorted(maps.Keys(m)) {
		for _, v := range m[k] {
			result = append(result, k, v)
		}
	}
	return result
}

// UnwrapSlice unwraps and copies a slice of strings from an untyped map.
func UnwrapSlice(v any) []string {
	var ss []string
	switch vv := v.(type) {
	case []string:
		ss = slices.Clone(vv)
	case []any:
		ss = make([]string, len(vv))
		for i := range vv {
			ss[i] = fmt.Sprint(vv[i])
		}
	}
	ss = slices.DeleteFunc(ss, func(s string) bool {
		return s == ""
	})
	return ss
}
