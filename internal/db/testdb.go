package db

import (
	"context"
	"os"
	"time"

	"github.com/earthboundkid/errorx/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
)

func CreateTestDatabase(dbURL string) (p *pgxpool.Pool, err error) {
	defer errorx.Trace(&err)

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(ctx, "drop database if exists almanack_test"); err != nil {
		return nil, err
	}
	if _, err = db.Exec(ctx, "create database almanack_test"); err != nil {
		return nil, err
	}
	newCfg := cfg.Copy()
	newCfg.ConnConfig.Database = "almanack_test"

	newDB, err := pgxpool.NewWithConfig(ctx, newCfg)
	if err != nil {
		return nil, err
	}
	conn, err := newDB.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	mg, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
	if err != nil {
		return nil, err
	}
	if err = mg.LoadMigrations(os.DirFS("../../sql/schema")); err != nil {
		return nil, err
	}
	if err = mg.Migrate(ctx); err != nil {
		return nil, err
	}
	return newDB, nil
}
