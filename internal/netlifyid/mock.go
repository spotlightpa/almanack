package netlifyid

import (
	"net/http"

	"github.com/spotlightpa/almanack/pkg/common"
)

type MockAuthService struct{}

var _ AuthService = MockAuthService{}

func (mas MockAuthService) AuthFromHeader(r *http.Request) (*http.Request, error) {
	return r, nil
}

func (mas MockAuthService) AuthFromCookie(r *http.Request) (*http.Request, error) {
	return r, nil
}

func (mas MockAuthService) HasRole(r *http.Request, role string) error {
	common.Logger.Printf("mock auth checking for role %q", role)
	if r.Header.Get("Authorization") != "" {
		return nil
	}
	if _, err := r.Cookie("nf_jwt"); err == nil {
		return nil
	}
	common.Logger.Printf("missing Authorization header/cookie")
	return nil
}
