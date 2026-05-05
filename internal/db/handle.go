// Package db contains sqlc models and queries and various tools for dealing with the database.
package db

import (
	"context"
	"time"

	"github.com/earthboundkid/errorx/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/sql/schema"
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

// DBTX returns a pgxpool.Pool wrapped in a logger.
func (h Handle) DBTX() DBTX {
	return logger{h.p}
}

// Queries is the sqlc wrapper for SQL queries and commands.
func (h Handle) Queries() *Queries {
	return New(h.DBTX())
}

// Tx executes its callback inside a transaction.
func (h Handle) Tx(ctx context.Context, o pgx.TxOptions, f func(*Queries) error) error {
	return pgx.BeginTxFunc(ctx, h.p, o, func(tx pgx.Tx) error {
		return f(h.Queries())
	})
}

// Migrate runs pending tern migrations against the underlying pool.
// It uses the same `schema_version` version table that the `tern migrate` CLI uses.
func (h Handle) Migrate(ctx context.Context) (err error) {
	defer errorx.Trace(&err)

	conn, err := h.p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	mg, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
	if err != nil {
		return err
	}

	if err = mg.LoadMigrations(schema.FS); err != nil {
		return err
	}
	return mg.Migrate(ctx)
}
