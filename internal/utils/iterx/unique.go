package iterx

import (
	"iter"

	"golang.org/x/exp/collection/set"
)

// UniquesFunc yields values in sequence but only once according to the keyfunc.
func UniquesFunc[T any, K comparable](seq iter.Seq[T], keyfunc func(T) K) iter.Seq[T] {
	return func(yield func(T) bool) {
		priorSet := set.Of[K]()
		for val := range seq {
			if priorSet.Insert(keyfunc(val)) && !yield(val) {
				return
			}
		}
	}
}
