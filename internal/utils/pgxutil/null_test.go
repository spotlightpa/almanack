package pgxutil_test

import (
	"testing"
	"unsafe"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/utils/pgxutil"
)

func TestNilSliceToEmpty(t *testing.T) {
	be :=
		// nil slice becomes non-nil empty slice
		assert.FailNow(t)

	var s []string
	be.True(s == nil)
	result := pgxutil.NilSliceToEmpty(s)
	be.
		True(result != nil).
		EqualLength(result, 0)

	// non-nil empty slice is returned as-is
	empty := []string{}
	be.
		EqualLength(pgxutil.NilSliceToEmpty(empty), 0).
		Equal(unsafe.SliceData(pgxutil.NilSliceToEmpty(empty)), unsafe.SliceData(empty))

	// populated slice is returned unchanged
	populated := []string{"a", "b"}
	be.SlicesEqual(pgxutil.NilSliceToEmpty(populated), populated)

	// works with other element types
	var ints []int
	be.
		True(pgxutil.NilSliceToEmpty(ints) != nil).
		EqualLength(pgxutil.NilSliceToEmpty(ints), 0)
}
