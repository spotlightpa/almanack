// Package slicex has slice helpers not in Go std slices.
package slicex

import (
	"slices"
)

// UniquesFunc deduplicates the slice by removing any items with the same key according to the keyfunc.
func UniquesFunc[S ~[]T, T any, K comparable](s *S, keyfunc func(T) K) {
	priorSet := make(map[K]struct{})
	*s = slices.DeleteFunc(*s, func(value T) bool {
		key := keyfunc(value)
		_, seen := priorSet[key]
		priorSet[key] = struct{}{}
		return seen
	})
}
