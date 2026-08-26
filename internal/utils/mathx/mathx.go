// Package mathx has convenience math functions.
package mathx

import "cmp"

// Clamp returns n clamped to [_min, _max].
func Clamp[N cmp.Ordered](n, _min, _max N) N {
	if _min > _max || _max < _min {
		panic("bad input to clamp")
	}
	return min(max(n, _min), _max)
}
