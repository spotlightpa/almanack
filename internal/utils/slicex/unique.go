// Package slicex has slice helpers not in Go std slices.
package slicex

import (
	"slices"

	"golang.org/x/exp/collection/set"
)

// UniquesFunc deduplicates the slice by removing any items with the same key according to the keyfunc.
func UniquesFunc[S ~[]T, T any, K comparable](s *S, keyfunc func(T) K) {
	priorSet := set.Of[K]()
	*s = slices.DeleteFunc(*s, func(value T) bool {
		return !priorSet.Insert(keyfunc(value))
	})
}
