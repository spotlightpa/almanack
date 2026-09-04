package integration_test

import (
	"strings"
	"testing"

	"github.com/earthboundkid/assert"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestRoles(t *testing.T) {
	be := assert.FailNow(t)
	dbhandle := createTestDB(t)
	q := dbhandle.Queries()

	ctx := t.Context()
	r := be.OK(q.UpsertRolesForAddress(ctx, db.UpsertRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        []string{"fooer"},
	}))
	be.
		Equal(r.EmailAddress, "a@foo.com").
		Equal(strings.Join(r.Roles, ","), "fooer")

	be.OK(q.UpsertRolesForDomain(ctx, db.UpsertRolesForDomainParams{
		Domain: "foo.com",
		Roles:  []string{"bar"},
	}))

	roles := be.OK(db.GetRolesForEmail(ctx, q, "a@foo.com"))
	be.Equal(strings.Join(roles, ","), "fooer")

	be.OK(q.UpsertRolesForAddress(ctx, db.UpsertRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        []string{},
	}))

	roles = be.OK(db.GetRolesForEmail(ctx, q, "a@foo.com"))
	be.Equal(strings.Join(roles, ","), "bar")

	be.OK(q.UpsertRolesForDomain(ctx, db.UpsertRolesForDomainParams{
		Domain: "foo.com",
		Roles:  []string{},
	}))

	roles = be.OK(db.GetRolesForEmail(ctx, q, "a@foo.com"))
	be.Equal(strings.Join(roles, ","), "")
}
