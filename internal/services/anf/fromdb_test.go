package anf_test

import (
	"path/filepath"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/anf"
)

func TestFromDB(t *testing.T) {
	testfile.Run(t, "testdata/*/item.json", func(t *testing.T, match string) {
		var item db.NewsFeedItem
		testfile.ReadJSON(t, match, &item)
		art := assert.FailNow(t).OK(anf.FromDB(&item))
		filename := filepath.Join(filepath.Dir(match), "article.json")
		testfile.EqualJSON(t, filename, art)
	})
}
