package db

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/emailx"
)

func GetRolesForEmail(ctx context.Context, q Querier, email string) (roles []string, err error) {
	// not likely to get pass Netlify with an invalid address, but why not check?
	if err = emailx.Validate(email); err != nil {
		return
	}
	_, domain := emailx.Split(email)
	if domain == "" {
		return nil, fmt.Errorf("invalid email: %q", email)
	}
	roles, err = q.GetRolesForAddress(ctx, email)
	if err != nil && !IsNotFound(err) {
		return
	}
	// if user has specific roles, early exit
	if err == nil && len(roles) > 0 {
		return
	}
	roles, err = q.GetRolesForDomain(ctx, domain)
	if err != nil && !IsNotFound(err) {
		return
	}
	// ignore any not found errors
	err = nil
	return
}
