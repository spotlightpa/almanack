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
	AuthFromHeader(r *http.Request) (*http.Request, error)
	AuthFromCookie(r *http.Request) (*http.Request, error)
	HasRole(r *http.Request, role string) (err error)
}

type ContentStore interface {
	GetFile(ctx context.Context, path string) (content string, err error)
	UpdateFile(ctx context.Context, msg, path string, content []byte) error
}

type EmailService interface {
	SendEmail(subject, body string) error
}

type Newsletter struct {
	Subject     string    `json:"subject"`
	Blurb       string    `json:"blurb"`
	Description string    `json:"description"`
	ArchiveURL  string    `json:"archive_url"`
	PublishedAt time.Time `json:"published_at"`
}

type NewletterService interface {
	ListNewletters(ctx context.Context, kind string) ([]Newsletter, error)
}
