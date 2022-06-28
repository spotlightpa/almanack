package syncx_test

import (
	"fmt"
	"sync"

	"github.com/spotlightpa/almanack/internal/syncx"
)

func ExampleOnce() {
	fourtytwo := 6 * 9
	var getMoL = syncx.Once(func() int {
		fmt.Println("calculating meaning of life...")
		fourtytwo -= 12
		return fourtytwo
	})
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			fmt.Println(getMoL())
			wg.Done()
		}()
	}
	wg.Wait()
	// Output:
	// calculating meaning of life...
	// 42
	// 42
	// 42
	// 42
	// 42
}