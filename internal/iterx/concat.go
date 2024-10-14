package iterx

import "iter"

// Concat streams each sequence it was passed in order.
func Concat[T any](seqs ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, seq := range seqs {
			for v := range seq {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Concat2 streams each sequence it was passed in order.
func Concat2[T1, T2 any](seqs ...iter.Seq2[T1, T2]) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		for _, seq := range seqs {
			for v1, v2 := range seq {
				if !yield(v1, v2) {
					return
				}
			}
		}
	}
}
