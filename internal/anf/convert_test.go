package anf_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/spotlightpa/almanack/internal/anf"
)

func TestConvert(t *testing.T) {
	testfile.Run(t, "testdata/*/article.html", func(t *testing.T, match string) {
		in := testfile.Read(t, match)
		art, err := anf.ConvertToAppleNews([]byte(in))
		be.NilErr(t, err)
		testfile.EqualJSON(t, testfile.Ext(match, ".json"), art)
	})
}
