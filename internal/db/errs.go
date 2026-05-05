package db

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/earthboundkid/resperr/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func NoRowsAs404(err error, format string, a ...any) error {
	if !IsNotFound(err) {
		return err
	}
	prefix := fmt.Sprintf(format, a...)
	return resperr.New(http.StatusNotFound, "%s: %w", prefix, err)
}

// IsUniquenessViolation tells whether the error was caused by violating a uniqueness constraint.
// If constraintName is not blank, it is also checked.
func IsUniquenessViolation(err error, constraintName string) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	return ok &&
		pgErr.Code == pgerrcode.UniqueViolation &&
		(constraintName == "" || pgErr.ConstraintName == constraintName)
}
