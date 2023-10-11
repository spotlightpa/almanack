// Package cmpx is a placeholder until Go 1.22
package cmpx

// Or is cmp.Or in Go 1.22
func Or[T comparable](vs ...T) T {
	var zero T
	for _, v := range vs {
		if v != zero {
			return v
		}
	}
	return zero
}
