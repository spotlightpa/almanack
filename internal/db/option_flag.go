package db

import (
	"context"
	"flag"

	"github.com/spotlightpa/almanack/pkg/common"
)

func FlagFromOption(ctx context.Context, q *Queries, fl *flag.FlagSet, name string) error {
	needsVal := true
	fl.Visit(func(f *flag.Flag) {
		if f.Name == name {
			needsVal = false
		}
	})
	if !needsVal {
		common.Logger.Printf("flag option: override of %s", name)
		return nil
	}
	common.Logger.Printf("flag option: getting %s", name)

	val, err := q.GetOption(ctx, name)
	if err != nil {
		return err
	}
	return fl.Set(name, val)
}
