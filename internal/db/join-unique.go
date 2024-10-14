package db

import "slices"

// ConcatUnique concatenates a and b then deduplicates
// the resulting slice by removing any items with the same key according to the keyfunc.
func ConcatUnique[T any, K comparable](a, b []T, keyfunc func(T) K) []T {
	r := slices.Concat(a, b)
	priorSet := make(map[K]struct{})
	r = slices.DeleteFunc(r, func(value T) bool {
		key := keyfunc(value)
		_, seen := priorSet[key]
		priorSet[key] = struct{}{}
		return seen
	})
	return r
}
