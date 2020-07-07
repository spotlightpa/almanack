package almanack

import (
	"context"
	"net/http"
)

var (
	BuildVersion string = "Development"
	DeployURL    string = "http://localhost"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type AuthService interface {
	AddToRequest(r *http.Request) (*http.Request, error)
	HasRole(r *http.Request, role string) (err error)
}

type ContentStore interface {
	GetFile(ctx context.Context, path string) (content string, err error)
	UpdateFile(ctx context.Context, msg, path string, content []byte) error
}

type EmailService interface {
	SendEmail(subject, body string) error
}

type DataStore interface {
	Get(key string, v interface{}) error
	Set(key string, v interface{}) error
	GetSet(key string, getv, setv interface{}) (err error)
	GetLock(key string) (unlock func(), err error)
}
