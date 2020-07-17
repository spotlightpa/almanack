package db

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/carlmjohnson/resperr"
)

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func ExpectNotFound(err error) error {
	if IsNotFound(err) {
		return resperr.WithStatusCode(err, http.StatusNotFound)
	}
	return err
}
