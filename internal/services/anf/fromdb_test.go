package anf_test

import (
	"github.com/earthboundkid/assert"
	"github.com/earthboundkid/assert/testfile"
	"github.com/spotlightpa/almanack/internal/db"
	"github.com/spotlightpa/almanack/internal/services/anf"
	"path/filepath"
	"testing"
)

func TestFromDB(t *testing.T) {
	be := assert.FailNow(t)
	testfile.Run(t, "testdata/*/item.json", func(t *testing.T, match string) {
		var item db.NewsFeedItem
		testfile.ReadJSON(t, match, &item)
		art, err := anf.FromDB(&item)
		be.Zero(err)
		filename := filepath.Join(filepath.Dir(match), "article.json")
		testfile.EqualJSON(t, filename, art)
	})
}
