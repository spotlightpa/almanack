package common

import (
	"context"
	"net/http"
)

type Logger interface {
	Printf(format string, v ...any)
}

type AuthService interface {
	AuthFromHeader(r *http.Request) (*http.Request, error)
	AuthFromCookie(r *http.Request) (*http.Request, error)
	HasRole(r *http.Request, role string) (err error)
}

type ContentStore interface {
	GetFile(ctx context.Context, path string) (content string, err error)
	UpdateFile(ctx context.Context, msg, path string, content []byte) error
}

type EmailService interface {
	SendEmail(ctx context.Context, subject, body string) error
}
