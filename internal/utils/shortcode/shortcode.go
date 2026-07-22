package shortcode

import (
	"html"
	"maps"
	"slices"
	"strings"
)

// New returns a new shortcode with escaped attrs.
// Panics if the number of attrs isn't even.
func New(tag string, attrs ...string) string {
	if len(attrs)%2 != 0 {
		panic("uneven attrs")
	}
	var sb strings.Builder
	sb.WriteString("{{<")
	sb.WriteString(tag)
	for n := range len(attrs) / 2 {
		key := attrs[n*2]
		value := attrs[(n*2)+1]
		sb.WriteByte(' ')
		sb.WriteString(escapeAttr(key))
		sb.WriteByte('=')
		sb.WriteByte('"')
		sb.WriteString(escapeAttr(value))
		sb.WriteByte('"')
	}
	sb.WriteString(">}}")
	return sb.String()
}

func escapeAttr(s string) string {
	attr := html.EscapeString(s)
	attr = strings.ReplaceAll(attr, "\n", "&#10;")
	return attr
}

// New returns a new shortcode with escaped attrs.
// Attrs are sorted by key.
func NewMapAttrs(tag string, attrs map[string][]string) string {
	attrslice := make([]string, 0, len(attrs)*2)
	for _, k := range slices.Sorted(maps.Keys(attrs)) {
		for _, v := range attrs[k] {
			attrslice = append(attrslice, k, v)
		}
	}
	return New(tag, attrslice...)
}
