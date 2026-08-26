// Package mathx has convenience math functions.
package mathx

import "cmp"

// Clamp returns n clamped to [_min, _max].
func Clamp[N cmp.Ordered](n, _min, _max N) N {
	return min(max(n, _min), _max)
}
