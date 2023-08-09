package db

import (
	"context"
	"flag"

	"github.com/spotlightpa/almanack/pkg/almlog"
)

func FlagFromOption(ctx context.Context, q *Queries, fl *flag.FlagSet, name string) error {
	l := almlog.FromContext(ctx)
	needsVal := true
	fl.Visit(func(f *flag.Flag) {
		if f.Name == name {
			needsVal = false
		}
	})
	if !needsVal {
		l.InfoContext(ctx, "db.FlagFromOption: override", "name", name)
		return nil
	}
	l.InfoContext(ctx, "db.FlagFromOption: get", "name", name)

	val, err := q.GetOption(ctx, name)
	if err != nil {
		return err
	}
	return fl.Set(name, val)
}
