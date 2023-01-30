package netlifyid

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/carlmjohnson/errorx"
	"github.com/carlmjohnson/resperr"
	"github.com/spotlightpa/almanack/pkg/almlog"
	"golang.org/x/exp/slog"
)

func NewService(isLambda bool) AuthService {
	if isLambda {
		return NetlifyAuth{}
	}
	almlog.Logger.Warn("mocking auth")
	return MockAuthService{}
}

type AuthService interface {
	AuthFromHeader(r *http.Request) (*http.Request, error)
	AuthFromCookie(r *http.Request) (*http.Request, error)
	HasRole(r *http.Request, role string) (err error)
}

type NetlifyAuth struct{}

var _ AuthService = NetlifyAuth{}

func (as NetlifyAuth) AuthFromHeader(r *http.Request) (*http.Request, error) {
	netID, err := FromLambdaContext(r.Context())
	if err != nil {
		l := almlog.FromContext(r.Context())
		l.Error("netlify.AuthFromHeader", err)
		return nil, err
	}
	return addJWTToRequest(netID, r), nil
}

func (as NetlifyAuth) AuthFromCookie(r *http.Request) (*http.Request, error) {
	netID, err := FromCookie(r)
	if err != nil {
		l := almlog.FromContext(r.Context())
		l.Error("netlify.AuthFromCookie", err)
		return nil, err
	}
	return addJWTToRequest(netID, r), nil
}

func (as NetlifyAuth) HasRole(r *http.Request, role string) error {
	l := almlog.FromContext(r.Context())
	if jwt := FromContext(r.Context()); jwt != nil {
		hasRole := jwt.HasRole(role)
		level := slog.LevelInfo
		if !hasRole {
			level = slog.LevelWarn
		}
		l.Log(level, "permission middleware",
			"requires-role", role,
			"has-role", hasRole,
		)
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

	err := resperr.WithUserMessage(
		fmt.Errorf("no user info provided: is this localhost?"),
		"no user info provided",
	)
	l.Error("netlify.HasRole: no identity found: running on AWS?", err)

	return err
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
	defer errorx.Trace(&err)

	c, err := r.Cookie("nf_jwt")
	if err != nil {
		return nil, err
	}
	defer func(value string) {
		if err != nil {
			err = fmt.Errorf("malformed cookie value %q: %w",
				value, err)
		}
	}(c.Value)
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
