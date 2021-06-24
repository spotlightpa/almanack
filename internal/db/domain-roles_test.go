package db_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/spotlightpa/almanack/internal/db"
)

func check(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func eq(t *testing.T, want, have string) {
	t.Helper()
	if want != have {
		t.Fatalf("want %q; have %q", want, have)
	}
}

func TestRoles(t *testing.T) {
	dbURL := os.Getenv("ALMANACK_TEST_DATABASE")
	if dbURL == "" {
		t.Skip("ALMANACK_TEST_DATABASE not set")
	}
	q, err := db.Open(dbURL)
	check(t, err, "could not open DB")
	ctx := context.Background()
	r, err := q.SetRolesForAddress(ctx, db.SetRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        []string{"fooer"},
	})
	check(t, err, "could not set address role")

	t.Cleanup(func() {
		q.SetRolesForAddress(ctx, db.SetRolesForAddressParams{
			EmailAddress: "a@foo.com",
			Roles:        []string{},
		})
	})
	eq(t, "a@foo.com", r.EmailAddress)
	eq(t, "fooer", strings.Join(r.Roles, ","))

	_, err = q.SetRolesForDomain(ctx, db.SetRolesForDomainParams{
		Domain: "foo.com",
		Roles:  []string{"bar"},
	})
	check(t, err, "set domain roles")

	t.Cleanup(func() {
		q.SetRolesForDomain(ctx, db.SetRolesForDomainParams{
			Domain: "foo.com",
			Roles:  []string{},
		})
	})

	roles, err := db.GetRolesForEmail(ctx, q, "a@foo.com")
	check(t, err, "get roles")
	eq(t, "fooer", strings.Join(roles, ","))

	_, err = q.SetRolesForAddress(ctx, db.SetRolesForAddressParams{
		EmailAddress: "a@foo.com",
		Roles:        []string{},
	})
	check(t, err, "get roles")

	roles, err = db.GetRolesForEmail(ctx, q, "a@foo.com")
	check(t, err, "get roles")
	eq(t, "bar", strings.Join(roles, ","))

	_, err = q.SetRolesForDomain(ctx, db.SetRolesForDomainParams{
		Domain: "foo.com",
		Roles:  []string{},
	})
	check(t, err, "set domain roles")

	roles, err = db.GetRolesForEmail(ctx, q, "a@foo.com")
	check(t, err, "get roles")
	eq(t, "", strings.Join(roles, ","))
}
