package netlifyid

import (
	"net/http"

	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/pkg/common"
)

type MockAuthService struct{ common.Logger }

var _ common.AuthService = MockAuthService{}

func (mas MockAuthService) AuthFromHeader(r *http.Request) (*http.Request, error) {
	return r, nil
}

func (mas MockAuthService) AuthFromCookie(r *http.Request) (*http.Request, error) {
	return r, nil
}

func (mas MockAuthService) HasRole(r *http.Request, role string) error {
	mas.Logger.Printf("mock auth checking for role %q", role)
	if r.Header.Get("Authorization") == "" {
		return resperr.New(http.StatusUnauthorized, "missing Authorization header")
	}
	return nil
}
