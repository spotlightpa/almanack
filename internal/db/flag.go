package db

import (
	"context"
	"database/sql"
	"flag"
	"time"
)

// FlagVar adds an option to the specified FlagSet (or flag.CommandLine if nil)
// that creates and tests a DB
func FlagVar(fl *flag.FlagSet, name, usage string) (q *Querier) {
	if fl == nil {
		fl = flag.CommandLine
	}
	q = new(Querier)
	fl.Func(name, usage, func(dbURL string) error {
		var err error
		*q, err = Open(dbURL)
		return err
	})
	return
}

func Open(dbURL string) (q Querier, err error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	q = New(db)
	return q, nil
}
