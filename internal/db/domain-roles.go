package db

import (
	"context"
	"fmt"

	"github.com/carlmjohnson/emailx"
)

func GetRolesForEmailDomain(ctx context.Context, q Querier, email string) (roles []string, err error) {
	_, domain := emailx.Split(email)
	if domain == "" {
		return nil, fmt.Errorf("invalid email: %q", email)
	}

	roles, err = q.GetRolesForDomain(ctx, domain)
	if err != nil && !IsNotFound(err) {
		return nil, err
	}
	return roles, nil
}
