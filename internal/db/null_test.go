package db_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestNilSliceToEmpty(t *testing.T) {
	// nil slice becomes non-nil empty slice
	var s []string
	be.True(t, s == nil)
	result := db.NilSliceToEmpty(s)
	be.True(t, result != nil)
	be.Equal(t, 0, len(result))

	// non-nil empty slice is returned as-is
	empty := []string{}
	be.True(t, db.NilSliceToEmpty(empty) != nil)
	be.Equal(t, 0, len(db.NilSliceToEmpty(empty)))

	// populated slice is returned unchanged
	populated := []string{"a", "b"}
	be.AllEqual(t, populated, db.NilSliceToEmpty(populated))

	// works with other element types
	var ints []int
	be.True(t, db.NilSliceToEmpty(ints) != nil)
	be.Equal(t, 0, len(db.NilSliceToEmpty(ints)))
}
