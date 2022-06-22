package syncx

import "sync"

func Once[T any](initializer func() T) func() T {
	var once sync.Once
	var t T
	f := func() {
		t = initializer()
	}
	return func() T {
		once.Do(f)
		return t
	}
}
