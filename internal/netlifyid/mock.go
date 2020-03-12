package netlifyid

import (
	"net/http"

	"github.com/spotlightpa/almanack/pkg/almanack"
	"github.com/spotlightpa/almanack/pkg/errutil"
)

type MockAuthService struct{ almanack.Logger }

var _ almanack.AuthService = MockAuthService{}

func (mas MockAuthService) AddToRequest(r *http.Request) (*http.Request, error) {
	return r, nil
}

func (mas MockAuthService) HasRole(r *http.Request, role string) error {
	mas.Logger.Printf("mock auth checking for role %q", role)
	if r.Header.Get("Authorization") == "" {
		return errutil.Unauthorized
	}
	return nil
}
