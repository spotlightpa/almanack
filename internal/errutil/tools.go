package errutil

import (
	"sync"

	"github.com/carlmjohnson/errutil"
)

var Merge = errutil.Merge

type Slice = errutil.Slice

func ExecParallel(fs ...func() error) error {
	var (
		size = len(fs)
		wg   sync.WaitGroup
		errs = make(Slice, size)
	)
	wg.Add(size)
	for i := range fs {
		go func(i int) {
			errs[i] = fs[i]()
			wg.Done()
		}(i)
	}
	wg.Wait()
	return errs.Merge()
}
