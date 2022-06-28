package db

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/spotlightpa/almanack/internal/stringx"
	"github.com/spotlightpa/almanack/pkg/common"
)

type logger struct {
	db DBTX
}

func (l logger) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	t, err := l.db.Exec(ctx, query, args...)
	l.log("Exec", time.Since(start))
	return t, err
}

func (l logger) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	rows, err := l.db.Query(ctx, query, args...)
	l.log("Query", time.Since(start))
	return rows, err
}

func (l logger) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	start := time.Now()
	row := l.db.QueryRow(ctx, query, args...)
	l.log("QueryRow", time.Since(start))
	return row
}

func (l logger) log(kind string, d time.Duration) {
	pc, file, line, ok := runtime.Caller(2)
	prefix := "unknown function"
	if ok {
		f := runtime.FuncForPC(pc)
		file = filepath.Base(file)
		_, name, _ := stringx.LastCut(f.Name(), ".")
		prefix = fmt.Sprintf("%s(%s:%d)", name, file, line)
	}
	common.Logger.Printf("%s %s in %v", kind, prefix, d)
}