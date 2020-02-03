package netlifyid

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spotlightpa/almanack/internal/errutil"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func NewService(isLambda bool, l almanack.Logger) almanack.AuthService {
	if isLambda {
		return AuthService{l}
	}
	return MockAuthService{l}
}

type AuthService struct{ almanack.Logger }

var _ almanack.AuthService = AuthService{}

type netlifyidContextType int

const netlifyidContextKey = iota

func (as AuthService) AddToRequest(r *http.Request) (*http.Request, error) {
	as.Logger.Printf("auth wrapping request")
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
		err := errutil.Response{
			StatusCode: http.StatusForbidden,
			Message:    http.StatusText(http.StatusForbidden),
			Log: fmt.Sprintf(
				"unauthorized user only had roles: %v",
				jwt.User.AppMetadata.Roles),
		}
		return err
	}
	as.Logger.Printf("no identity found: running on AWS?")
	err := errutil.Response{
		StatusCode: http.StatusInternalServerError,
		Message:    "user info not set",
		Log:        "no user info: is this localhost?",
	}
	return err
}
