package db

import (
	"context"
	"fmt"
	"strings"
)

func domain(email string) string {
	if i := strings.LastIndexByte(email, '@'); i != -1 {
		return email[i+1:]
	}

	return ""
}

func GetRolesForEmailDomain(ctx context.Context, q Querier, email string) (roles []string, err error) {
	domain := domain(email)
	if domain == "" {
		return nil, fmt.Errorf("invalid email: %q", email)
	}

	roles, err = q.GetRolesForDomain(ctx, domain)
	if err != nil && !IsNotFound(err) {
		return nil, err
	}
	return roles, nil
}
