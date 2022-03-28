package netlifyid

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/carlmjohnson/errutil"
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

func (as AuthService) AuthFromHeader(r *http.Request) (*http.Request, error) {
	netID, err := FromLambdaContext(r.Context())
	if err != nil {
		as.Logger.Printf("could not wrap request: %v", err)
		return nil, err
	}
	return addJWTToRequest(netID, r), nil
}

func (as AuthService) AuthFromCookie(r *http.Request) (*http.Request, error) {
	netID, err := FromCookie(r)
	if err != nil {
		as.Logger.Printf("could not wrap request: %v", err)
		return nil, err
	}
	return addJWTToRequest(netID, r), nil
}

func (as AuthService) HasRole(r *http.Request, role string) error {
	if jwt := FromContext(r.Context()); jwt != nil {
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

func FromLambdaContext(ctx context.Context) (*JWT, error) {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok {
		return nil, resperr.WithUserMessage(
			fmt.Errorf("no context given: is this localhost?"),
			"no context provided",
		)
	}
	encoded := lc.ClientContext.Custom["netlify"]
	jwtBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("could not decode context %q: %v", encoded, err)
	}
	var netID JWT
	if err = json.Unmarshal(jwtBytes, &netID); err != nil {
		return nil, fmt.Errorf("could not decode identity %q: %v", encoded, err)
	}
	return &netID, nil
}

// Danger! Does not verify JWT! Do not use in insecure context.
func FromCookie(r *http.Request) (id *JWT, err error) {
	defer errutil.Prefix(&err, "problem retrieving JWT from cookie")

	c, err := r.Cookie("nf_jwt")
	if err != nil {
		return nil, err
	}
	defer errutil.Prefix(&err, fmt.Sprintf("malformed cookie value: %q", c.Value))
	_, s, ok := strings.Cut(c.Value, ".")
	if !ok {
		return nil, fmt.Errorf("missing initial dot")
	}
	s, _, ok = strings.Cut(s, ".")
	if !ok {
		return nil, fmt.Errorf("missing second dot")
	}
	jwtBytes, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var user User
	if err = json.Unmarshal(jwtBytes, &user); err != nil {
		return nil, err
	}

	return &JWT{User: user}, nil
}
