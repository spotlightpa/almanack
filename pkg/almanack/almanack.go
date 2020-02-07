package almanack

import (
	"context"
	"net/http"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type AuthService interface {
	AddToRequest(r *http.Request) (*http.Request, error)
	HasRole(r *http.Request, role string) (err error)
}

type ContentStore interface {
	CreateFile(ctx context.Context, msg, path string, content []byte) error
}

type EmailService interface {
	SendEmail(subject, body string) error
}

type ImageStore interface {
	GetSignedUpload() (signedURL, filename string, err error)
}

type DataStore interface {
	Get(key string, v interface{}) error
	Set(key string, v interface{}) error
	GetSet(key string, getv, setv interface{}) (err error)
	GetLock(key string) (unlock func(), err error)
}
