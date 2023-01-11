package db

import (
	"context"
	"flag"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spotlightpa/almanack/pkg/almlog"
)

// AddFlags adds an option to the specified FlagSet that creates and tests a DB
func AddFlags(fl *flag.FlagSet, name, usage string) (q *Queries, tx *Txable) {
	q = new(Queries)
	tx = new(Txable)
	fl.Func(name, usage, func(dbURL string) error {
		p, err := Open(dbURL)
		if p != nil {
			q.db = logger{p}
			tx.p = p
		}
		return err
	})
	return
}

func Open(dbURL string) (p *pgxpool.Pool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	db, err := pgxpool.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	go func() {
		ctx2, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := db.Ping(ctx2); err != nil {
			almlog.Slogger.Error("db.Ping", err)
		}
	}()

	return db, nil
}

type Txable struct {
	p *pgxpool.Pool
}

func (txable Txable) Begin(ctx context.Context, o pgx.TxOptions, f func(*Queries) error) error {
	return txable.p.BeginTxFunc(ctx, o, func(tx pgx.Tx) error {
		return f(New(logger{tx}))
	})
}
