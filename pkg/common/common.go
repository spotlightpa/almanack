package common

import (
	"context"
	"net/http"
	"time"
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

type FileStore interface {
	GetSignedURL(srcPath string, h http.Header) (signedURL string, err error)
	BuildURL(srcPath string) string
}

type Newsletter struct {
	Subject     string    `json:"subject"`
	ArchiveURL  string    `json:"archive_url"`
	PublishedAt time.Time `json:"published_at"`
}

type NewletterService interface {
	ListNewletters(ctx context.Context, kind string) ([]Newsletter, error)
}
