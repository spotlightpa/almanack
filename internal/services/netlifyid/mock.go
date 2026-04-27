package netlifyid

import (
	"net/http"

	"github.com/spotlightpa/almanack/pkg/almlog"
)

type MockAuthService struct{}

var _ AuthService = MockAuthService{}

func (mas MockAuthService) AuthFromHeader(r *http.Request) (*http.Request, error) {
	r = addJWTToRequest(&JWT{User: User{Email: "mock"}}, r)
	return r, nil
}

func (mas MockAuthService) AuthFromCookie(r *http.Request) (*http.Request, error) {
	return r, nil
}

func (mas MockAuthService) HasRole(r *http.Request, role string) error {
	l := almlog.FromContext(r.Context())
	l.DebugContext(r.Context(), "mock permission middleware",
		"requires-role", role,
		"has-role", true)

	if r.Header.Get("Authorization") != "" {
		return nil
	}
	if _, err := r.Cookie("nf_jwt"); err == nil {
		return nil
	}
	l.WarnContext(r.Context(), "missing Authorization header/cookie")
	return nil
}
