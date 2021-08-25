package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/resperr"
)

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func NoRowsAs404(err error, format string, a ...interface{}) error {
	if !IsNotFound(err) {
		return err
	}
	prefix := fmt.Sprintf(format, a...)
	return resperr.New(http.StatusNotFound, "%s: %w", prefix, err)
}
