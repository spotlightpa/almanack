// Package iterx has iteration utilities.
package iterx

import (
	"iter"
)

// Filter returns a sequence of matching items.
func Filter[T any](seq iter.Seq[T], match func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if match(v) && !yield(v) {
				return
			}
		}
	}
}

// First returns the first item in a sequence or the zero value.
func First[T any](seq iter.Seq[T]) (v T) {
	for v := range seq {
		return v
	}
	return
}
