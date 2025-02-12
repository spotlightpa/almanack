package stringx_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
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

func TestExtractName(t *testing.T) {
	rt := be.Relaxed(t)

	type testcase struct {
		Input string
		Want  []string
	}
	// Manual cases
	cases := []testcase{
		{"", []string{}},
		{"Spotlight PA Staff", []string{}},
		{"John Stafford", []string{"John Stafford"}},
		{
			"Stephen Caruso of Spotlight PA, Kate Huangpu of Spotlight PA, and Katie Meyer of Spotlight PA",
			[]string{"Stephen Caruso", "Kate Huangpu", "Katie Meyer"},
		},
		{"Kate Huangpu y Elizabeth Estrada de Spotlight PA", []string{"Kate Huangpu", "Elizabeth Estrada"}},
		{
			"STEPHEN CARUSO OF SPOTLIGHT PA, KATE HUANGPU OF SPOTLIGHT PA, AND KATIE MEYER OF SPOTLIGHT PA",
			[]string{"STEPHEN CARUSO", "KATE HUANGPU", "KATIE MEYER"},
		},
		{"Jane Often", []string{"Jane Often"}},
		{"Susana González-Lopez", []string{"Susana González-Lopez"}},
		{"Andy Fortune", []string{"Andy Fortune"}},
		{"Samuel O’Neal for Spotlight PA", []string{"Samuel O’Neal"}},
		{
			"Carter Walker of Votebeat and Laura Benshoff for Votebeat",
			[]string{"Carter Walker", "Laura Benshoff"},
		},
		{"Wyatt Massey of Spotlight PA State College", []string{"Wyatt Massey"}},
		{"Danielle Ohl of Spotlight PA; Jessica Lussenhop of ProPublica; and Irina Bucur, Tracy Leturgey and Eddie Trizzino of The Butler Eagle", []string{"Danielle Ohl", "Jessica Lussenhop", "Irina Bucur", "Tracy Leturgey", "Eddie Trizzino"}},
		{"Amanda Fries / For Spotlight PA", []string{"Amanda Fries"}},
	}
	for _, tc := range cases {
		be.AllEqual(rt, tc.Want, stringx.ExtractNames(tc.Input))
	}

	// Bulk cases
	testfile.ReadJSON(t, "testdata/extract-name-cases.json", &cases)
	for _, tc := range cases {
		be.AllEqual(rt, tc.Want, stringx.ExtractNames(tc.Input))
	}
}
