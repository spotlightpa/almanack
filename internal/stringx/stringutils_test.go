package stringx_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/stringx"
)

func TestSlugifyURL(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"", ""},
		{"  b  ", "b"},
		{"  ab  ", "ab"},
		{"  a b the c  ", "b-c"},
		{"Pa.'s favorite", "pennsylvanias-favorite"},
		{"the (fort~Nightly)   news  ", "fort-nightly-news"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			be.Equal(t, tc.want, stringx.SlugifyURL(tc.input))
		})
	}
}

func TestSlugifyFilename(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"", ""},
		{"  b  ", "-b-"},
		{"  ab  ", "-ab-"},
		{"  a b the c  ", "-a-b-the-c-"},
		{"Pa.'s favorite", "pa.-s-favorite"},
		{"the (fort~Nightly)   news  ", "the-fort-nightly-news-"},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			be.Equal(t, tc.want, stringx.SlugifyFilename(tc.input))
		})
	}
}

func TestRemoveParens(t *testing.T) {
	// Test cases
	testCases := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"No parentheses", "No parentheses"},
		{"a(b", "a"},
		{"a(b))c", "ac"},
		{"0)))1((()))2", "012"},
		{"(Welcome) to (OpenAI)", " to "},
		{"(Nested (parentheses) in) the string", " the string"},
	}

	// Run test cases
	for _, tc := range testCases {
		be.Equal(t, tc.want, stringx.RemoveParens(tc.input))
	}
}
