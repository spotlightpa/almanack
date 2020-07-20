package netlifyid

import (
	"context"
	"fmt"
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/pkg/common"
)

func NewService(isLambda bool, l common.Logger) common.AuthService {
	if isLambda {
		return AuthService{l}
	}
	l.Printf("mocking auth")
	return MockAuthService{l}
}

type AuthService struct{ common.Logger }

var _ common.AuthService = AuthService{}

type netlifyidContextType int

const netlifyidContextKey = iota

func (as AuthService) AddToRequest(r *http.Request) (*http.Request, error) {
	netID, err := FromRequest(r)
	if err != nil {
		as.Logger.Printf("could not wrap request: %v", err)
		return nil, err
	}
	ctx := context.WithValue(r.Context(), netlifyidContextKey, netID)
	return r.WithContext(ctx), nil
}

func jwtFromRequest(r *http.Request) *JWT {
	ctx := r.Context()
	val, _ := ctx.Value(netlifyidContextKey).(*JWT)
	return val
}

func (as AuthService) HasRole(r *http.Request, role string) error {
	if jwt := jwtFromRequest(r); jwt != nil {
		hasRole := jwt.HasRole(role)
		as.Logger.Printf("permission middleware: %s has role %s == %t",
			jwt.User.Email, role, hasRole)
		if hasRole {
			return nil
		}

		return resperr.New(
			http.StatusUnauthorized,
			"unauthorized user %s only had roles: %v",
			jwt.User.Email,
			jwt.User.AppMetadata.Roles,
		)
	}
	as.Logger.Printf("no identity found: running on AWS?")

	return resperr.WithUserMessage(
		fmt.Errorf("no user info provided: is this localhost?"),
		"no user info provided",
	)
}
