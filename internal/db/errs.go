package db

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/jackc/pgx/v4"
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
