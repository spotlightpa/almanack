package iterx

import (
	"iter"
)

// UniquesFunc yields values in sequence but only once according to the keyfunc.
func UniquesFunc[T any, K comparable](seq iter.Seq[T], keyfunc func(T) K) iter.Seq[T] {
	return func(yield func(T) bool) {
		priorSet := make(map[K]struct{})
		for val := range seq {
			key := keyfunc(val)
			_, seen := priorSet[key]
			if seen {
				continue
			}
			priorSet[key] = struct{}{}
			if !yield(val) {
				return
			}
		}
	}
}
