package integration_test

import (
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/almlog"
	"github.com/spotlightpa/almanack/internal/utils/pgxutil"
)

func TestIsUniquenessViolation(t *testing.T) {
	be := assert.FailNow(t)
	almlog.UseTestLogger(t)
	dbhandle := createTestDB(t)
	dbtx := dbhandle.DBTX()
	{ // No errors to insert some key
		_, err := dbtx.Exec(t.Context(), "insert into option(key, value) values ('k', 'v')")
		be.
			Zero(err).
			False(pgxutil.IsUniquenessViolation(err, "")).
			False(pgxutil.IsUniquenessViolation(err, "blah")).
			False(pgxutil.IsUniquenessViolation(err, "option_key_key"))
	}
	{ // Get option_key_key uniqueness errors on repeat insertions of the same key
		_, err := dbtx.Exec(t.Context(), "insert into option(key, value) values ('k', 'v')")
		be.
			NotZero(err).
			True(pgxutil.IsUniquenessViolation(err, "")).
			False(pgxutil.IsUniquenessViolation(err, "blah")).
			True(pgxutil.IsUniquenessViolation(err, "option_key_key"))
	}
}
