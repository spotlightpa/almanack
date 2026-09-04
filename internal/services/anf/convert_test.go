package anf_test

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/services/anf"
)

func TestConvert(t *testing.T) {
	testfile.Run(t, "testdata/*/article.html", func(t *testing.T, match string) {
		in := testfile.Read(t, match)
		art := assert.FailNow(t).OK(anf.ConvertToAppleNews(in, "http://www.spotlightpa.org"))
		testfile.EqualJSON(t, testfile.Ext(match, ".json"), art)
	})
}
