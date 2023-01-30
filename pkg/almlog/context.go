package almlog

import (
	"context"

	"golang.org/x/exp/slog"
)

type contextKey struct{}

// NewContext returns a context that contains the given Logger.
// Use FromContext to retrieve the Logger.
func NewContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

// FromContext returns the Logger stored in ctx by NewContext, or the default
// Logger if there is none.
func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return l
	}
	return Logger
}

// Ctx retrieves a Logger from the given context using FromContext. Then it adds
// the given context to the Logger using WithContext and returns the result.
func Ctx(ctx context.Context) *slog.Logger {
	return FromContext(ctx).WithContext(ctx)
}
