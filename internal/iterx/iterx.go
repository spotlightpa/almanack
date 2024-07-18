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

// Concat2 streams seq1 then seq2
func Concat2[T1, T2 any](seq1, seq2 iter.Seq2[T1, T2]) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		for v1, v2 := range seq1 {
			if !yield(v1, v2) {
				return
			}
		}
		for v1, v2 := range seq2 {
			if !yield(v1, v2) {
				return
			}
		}
	}
}
