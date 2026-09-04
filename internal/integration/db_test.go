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
		be.Zero(err)
		be.False(pgxutil.IsUniquenessViolation(err, ""))
		be.False(pgxutil.IsUniquenessViolation(err, "blah"))
		be.False(pgxutil.IsUniquenessViolation(err, "option_key_key"))
	}
	{ // Get option_key_key uniqueness errors on repeat insertions of the same key
		_, err := dbtx.Exec(t.Context(), "insert into option(key, value) values ('k', 'v')")
		be.NotZero(err)
		be.True(pgxutil.IsUniquenessViolation(err, ""))
		be.False(pgxutil.IsUniquenessViolation(err, "blah"))
		be.True(pgxutil.IsUniquenessViolation(err, "option_key_key"))
	}
}
