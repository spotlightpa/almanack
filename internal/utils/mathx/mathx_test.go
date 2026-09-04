package mathx_test

import (
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/utils/mathx"
	"testing"
)

func TestClamp(t *testing.T) {
	be := assert.FailNow(t)
	for _, tc := range []struct {
		n, min, max, want int
	}{
		{5, 1, 10, 5},   // within range
		{0, 1, 10, 1},   // below min
		{99, 1, 10, 10}, // above max
		{1, 1, 10, 1},   // at min boundary
		{10, 1, 10, 10}, // at max boundary
		{10, 1, 1, 1},   // weird clamp inputs
	} {
		be.Equal(mathx.Clamp(tc.n, tc.min, tc.max), tc.want)
	}
	be.NotZero(assert.Catch(func() {
		mathx.Clamp(10, 100, 1)
	}))
}
