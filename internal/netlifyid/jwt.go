package netlifyid

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/carlmjohnson/resperr"
)

type JWT struct {
	Identity Identity `json:"identity"`
	SiteURL  string   `json:"site_url"`
	User     User     `json:"user"`
}

const adminRole = "admin"

func (jwt *JWT) HasRole(role string) bool {
	if jwt == nil {
		return false
	}
	for _, r := range jwt.User.AppMetadata.Roles {
		if r == role || r == adminRole {
			return true
		}
	}
	return false
}

func (jwt *JWT) Email() string {
	if jwt == nil {
		return ""
	}

	return jwt.User.Email
}

func (jwt *JWT) Username() string {
	if jwt == nil {
		return ""
	}

	return jwt.User.UserMetadata.FullName
}

type Identity struct {
	Token string `json:"token"`
	URL   string `json:"url"`
}
type AppMetadata struct {
	Provider string   `json:"provider"`
	Roles    []string `json:"roles"`
}
type UserMetadata struct {
	FullName string `json:"full_name"`
}
type User struct {
	AppMetadata  AppMetadata  `json:"app_metadata"`
	Email        string       `json:"email"`
	Exp          int          `json:"exp"`
	Sub          string       `json:"sub"`
	UserMetadata UserMetadata `json:"user_metadata"`
}

func FromRequest(r *http.Request) (*JWT, error) {
	return FromContext(r.Context())
}

func FromContext(ctx context.Context) (*JWT, error) {
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
