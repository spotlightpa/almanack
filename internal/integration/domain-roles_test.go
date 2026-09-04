package integration_test

import (
	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/db"
	"strings"
	"testing"
)

func TestRoles(t *testing.T) {
	be := assert.FailNow(t)
	dbhandle := createTestDB(t)
	q := dbhandle.Queries()

	ctx := t.Context()
	r, err := q.UpsertRolesForAddress(ctx, db.UpsertRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        []string{"fooer"},
	})
	be.Zero(err)

	be.Equal(r.EmailAddress, "a@foo.com")
	be.Equal(strings.Join(r.Roles, ","), "fooer")

	_, err = q.UpsertRolesForDomain(ctx, db.UpsertRolesForDomainParams{
		Domain: "foo.com",
		Roles:  []string{"bar"},
	})
	be.Zero(err)

	roles, err := db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.Zero(err)
	be.Equal(strings.Join(roles, ","), "fooer")

	_, err = q.UpsertRolesForAddress(ctx, db.UpsertRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        []string{},
	})
	be.Zero(err)

	roles, err = db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.Zero(err)
	be.Equal(strings.Join(roles, ","), "bar")

	_, err = q.UpsertRolesForDomain(ctx, db.UpsertRolesForDomainParams{
		Domain: "foo.com",
		Roles:  []string{},
	})
	be.Zero(err)

	roles, err = db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.Zero(err)
	be.Equal(strings.Join(roles, ","), "")
}
