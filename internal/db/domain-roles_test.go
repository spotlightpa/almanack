package db_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/carlmjohnson/be"
	"github.com/spotlightpa/almanack/internal/db"
)

func TestRoles(t *testing.T) {
	dbURL := os.Getenv("ALMANACK_TEST_DATABASE")
	if dbURL == "" {
		t.Skip("ALMANACK_TEST_DATABASE not set")
	}
	p, err := db.Open(dbURL)
	be.NilErr(t, err)
	q := db.New(p)
	ctx := context.Background()
	r, err := q.SetRolesForAddress(ctx, db.SetRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        []string{"fooer"},
	})
	be.NilErr(t, err)

	t.Cleanup(func() {
		q.SetRolesForAddress(ctx, db.SetRolesForAddressParams{
			EmailAddress: "a@foo.com",
			Roles:        []string{},
		})
	})
	be.Equal(t, "a@foo.com", r.EmailAddress)
	be.Equal(t, "fooer", strings.Join(r.Roles, ","))

	_, err = q.SetRolesForDomain(ctx, db.SetRolesForDomainParams{
		Domain: "foo.com",
		Roles:  []string{"bar"},
	})
	be.NilErr(t, err)

	t.Cleanup(func() {
		q.SetRolesForDomain(ctx, db.SetRolesForDomainParams{
			Domain: "foo.com",
			Roles:  []string{},
		})
	})

	roles, err := db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.NilErr(t, err)
	be.Equal(t, "fooer", strings.Join(roles, ","))

	_, err = q.SetRolesForAddress(ctx, db.SetRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        []string{},
	})
	be.NilErr(t, err)

	roles, err = db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.NilErr(t, err)
	be.Equal(t, "bar", strings.Join(roles, ","))

	_, err = q.SetRolesForDomain(ctx, db.SetRolesForDomainParams{
		Domain: "foo.com",
		Roles:  []string{},
	})
	be.NilErr(t, err)

	roles, err = db.GetRolesForEmail(ctx, q, "a@foo.com")
	be.NilErr(t, err)
	be.Equal(t, "", strings.Join(roles, ","))
}
