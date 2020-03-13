package db

import (
	"database/sql"
	"errors"

	"github.com/spotlightpa/almanack/pkg/errutil"
)

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func ExpectNotFound(err error) error {
	if IsNotFound(err) {
		return errutil.NotFound
	}
	return err
}
