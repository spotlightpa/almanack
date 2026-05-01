// Package lazy has sync.Once wrappers
package lazy

import (
	"regexp"
	"sync"
)

// RE is a lazy regular expression.
func RE(str string) func() *regexp.Regexp {
	return sync.OnceValue(func() *regexp.Regexp {
		return regexp.MustCompile(str)
	})
}
