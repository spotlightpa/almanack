package anf_test

import (
	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/services/anf"
	"testing"
)

func TestConvert(t *testing.T) {
	be := assert.FailNow(t)
	testfile.Run(t, "testdata/*/article.html", func(t *testing.T, match string) {
		in := testfile.Read(t, match)
		art, err := anf.ConvertToAppleNews(in, "http://www.spotlightpa.org")
		be.Zero(err)
		testfile.EqualJSON(t, testfile.Ext(match, ".json"), art)
	})
}
