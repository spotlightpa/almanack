package db

import (
	"context"
	"flag"

	"golang.org/x/exp/slog"
)

func FlagFromOption(ctx context.Context, q *Queries, fl *flag.FlagSet, name string) error {
	l := slog.FromContext(ctx)
	needsVal := true
	fl.Visit(func(f *flag.Flag) {
		if f.Name == name {
			needsVal = false
		}
	})
	if !needsVal {
		l.Info("db.FlagFromOption: override", "name", name)
		return nil
	}
	l.Info("db.FlagFromOption: get", "name", name)

	val, err := q.GetOption(ctx, name)
	if err != nil {
		return err
	}
	return fl.Set(name, val)
}
