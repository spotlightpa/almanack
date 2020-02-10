package netlifyid

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/lambdacontext"

	"github.com/spotlightpa/almanack/pkg/errutil"
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
	lc, ok := lambdacontext.FromContext(r.Context())
	if !ok {
		return nil, errutil.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "missing context",
			Log:        "no context given: is this localhost?",
		}
	}
	encoded := lc.ClientContext.Custom["netlify"]
	jwtBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errutil.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "could not decode context",
			Log:        fmt.Sprintf("could not decode context %q", encoded),
		}
	}
	var netID JWT
	if err = json.Unmarshal(jwtBytes, &netID); err != nil {
		return nil, errutil.Response{
			StatusCode: http.StatusInternalServerError,
			Message:    "could not decode identity",
			Log:        fmt.Sprintf("could not decode identity %q: %v", encoded, err),
		}
	}
	return &netID, nil
}
