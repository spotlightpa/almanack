package db_test

import (
	"testing"
	"unsafe"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestNilSliceToEmpty(t *testing.T) {
	// nil slice becomes non-nil empty slice
	var s []string
	be.True(t, s == nil)
	result := db.NilSliceToEmpty(s)
	be.True(t, result != nil)
	be.EqualLength(t, 0, result)

	// non-nil empty slice is returned as-is
	empty := []string{}
	be.EqualLength(t, 0, db.NilSliceToEmpty(empty))
	be.Equal(t, unsafe.SliceData(empty), unsafe.SliceData(db.NilSliceToEmpty(empty)))

	// populated slice is returned unchanged
	populated := []string{"a", "b"}
	be.AllEqual(t, populated, db.NilSliceToEmpty(populated))

	// works with other element types
	var ints []int
	be.True(t, db.NilSliceToEmpty(ints) != nil)
	be.EqualLength(t, 0, db.NilSliceToEmpty(ints))
}
