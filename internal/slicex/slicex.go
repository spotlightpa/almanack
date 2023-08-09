// Package slicex has slice helpers that aren't in std slices.
package slicex

func DeleteFunc[T any, S ~[]T](sp *S, del func(T) bool) {
	filtered := (*sp)[:0]
	for _, v := range *sp {
		if !del(v) {
			filtered = append(filtered, v)
		}
	}
	*sp = filtered
	// TODO: clear tail once clear is in language
}
