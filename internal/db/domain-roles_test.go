package db_test

import (
	"context"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestRoles(t *testing.T) {
	p := createTestDB(t)

	q := db.New(p)
	ctx := context.Background()
	r, err := q.UpsertRolesForAddress(ctx, db.UpsertRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        db.Array("fooer"),
	})
	be.NilErr(t, err)

	be.Equal(t, "a@foo.com", r.EmailAddress)
	be.Equal(t, "fooer", strings.Join(r.Roles.Elements, ","))

	_, err = q.UpsertRolesForDomain(ctx, db.UpsertRolesForDomainParams{
		Domain: "foo.com",
		Roles:  db.Array("bar"),
	})
	be.NilErr(t, err)

	roles, err := db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.NilErr(t, err)
	be.Equal(t, "fooer", strings.Join(roles, ","))

	_, err = q.UpsertRolesForAddress(ctx, db.UpsertRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        pgtype.Array[string]{},
	})
	be.NilErr(t, err)

	roles, err = db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.NilErr(t, err)
	be.Equal(t, "bar", strings.Join(roles, ","))

	_, err = q.UpsertRolesForDomain(ctx, db.UpsertRolesForDomainParams{
		Domain: "foo.com",
		Roles:  pgtype.Array[string]{},
	})
	be.NilErr(t, err)

	roles, err = db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.NilErr(t, err)
	be.Equal(t, "", strings.Join(roles, ","))
}
