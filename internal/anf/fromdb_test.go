package anf_test

import (
	"path/filepath"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/carlmjohnson/be/testfile"
	"github.com/spotlightpa/almanack/internal/anf"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestFromDB(t *testing.T) {
	testfile.Run(t, "testdata/*/item.json", func(t *testing.T, match string) {
		var item db.AppleNewsFeed
		testfile.ReadJSON(t, match, &item)
		art, err := anf.FromDB(item)
		be.NilErr(t, err)
		filename := filepath.Join(filepath.Dir(match), "article.json")
		testfile.EqualJSON(t, filename, art)
	})
}
