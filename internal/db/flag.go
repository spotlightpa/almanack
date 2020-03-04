package db

import (
	"context"
	"database/sql"
	"flag"
	"time"
)

// FlagVar adds an option to the specified FlagSet (or flag.CommandLine if nil)
// that creates and tests a DB
func FlagVar(fl *flag.FlagSet, name, usage string) func() (q Querier, err error) {
	if fl == nil {
		fl = flag.CommandLine
	}
	dbURL := fl.String(name, "", usage)
	return func() (q Querier, err error) {
		return Open(*dbURL)
	}
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
