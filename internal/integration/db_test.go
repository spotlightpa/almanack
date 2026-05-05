package integration_test

import (
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestIsUniquenessViolation(t *testing.T) {
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	dbtx := dbhandle.DBTX()
	{ // No errors to insert some key
		_, err := dbtx.Exec(t.Context(), "insert into option(key, value) values ('k', 'v')")
		be.NilErr(t, err)
		be.False(t, db.IsUniquenessViolation(err, ""))
		be.False(t, db.IsUniquenessViolation(err, "blah"))
		be.False(t, db.IsUniquenessViolation(err, "option_key_key"))
	}
	{ // Get option_key_key uniqueness errors on repeat insertions of the same key
		_, err := dbtx.Exec(t.Context(), "insert into option(key, value) values ('k', 'v')")
		be.Nonzero(t, err)
		be.True(t, db.IsUniquenessViolation(err, ""))
		be.False(t, db.IsUniquenessViolation(err, "blah"))
		be.True(t, db.IsUniquenessViolation(err, "option_key_key"))
	}
}
