package common

import (
	"context"
)

type Logger interface {
	Printf(format string, v ...any)
}

type ContentStore interface {
	GetFile(ctx context.Context, path string) (content string, err error)
	UpdateFile(ctx context.Context, msg, path string, content []byte) error
}

type EmailService interface {
	SendEmail(ctx context.Context, subject, body string) error
}
