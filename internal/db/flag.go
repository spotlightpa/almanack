package db

import (
	"context"
	"database/sql"
	"flag"
	"time"
)

// FlagVar adds an option to the specified FlagSet (or flag.CommandLine if nil)
// that creates and tests a DB
func FlagVar(fl *flag.FlagSet, name, usage string) (q *Queries) {
	if fl == nil {
		fl = flag.CommandLine
	}
	q = new(Queries)
	fl.Func(name, usage, func(dbURL string) error {
		q2, err := Open(dbURL)
		*q = *q2
		return err
	})
	return
}

func Open(dbURL string) (q *Queries, err error) {
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
