package anf_test

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/services/anf"
)

func TestConvert(t *testing.T) {
	testfile.Run(t, "testdata/*/article.html", func(be assert.TB, match string) {
		in := testfile.Read(be, match)
		art := be.OK(anf.ConvertToAppleNews(in, "http://www.spotlightpa.org"))
		testfile.EqualJSON(be, testfile.Ext(match, ".json"), art)
	})
}
