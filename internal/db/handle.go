// Package db contains sqlc models and queries and various tools for dealing with the database.
package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spotlightpa/almanack/internal/almlog"
)

func Open(dbURL string) (p *pgxpool.Pool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	go func() {
		ctx2, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := db.Ping(ctx2); err != nil {
			almlog.Logger.Error("db.Ping", "err", err)
		}
	}()

	return db, nil
}

func NewHandle(p *pgxpool.Pool) *Handle {
	return &Handle{p}
}

// Handle wraps a pgxpool.Pool and returns logged Queries and transactions.
type Handle struct {
	p *pgxpool.Pool
}

func (h Handle) Pool() *pgxpool.Pool {
	return h.p
}

func (h Handle) Queries() *Queries {
	return &Queries{logger{h.p}}
}

func (h Handle) Begin(ctx context.Context, o pgx.TxOptions, f func(*Queries) error) error {
	return pgx.BeginTxFunc(ctx, h.p, o, func(tx pgx.Tx) error {
		return f(New(logger{tx}))
	})
}
