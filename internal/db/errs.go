package db

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/earthboundkid/resperr/v2"
	"github.com/jackc/pgx/v5"
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
