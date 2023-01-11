package netlifyid

import (
	"context"
	"net/http"

	"golang.org/x/exp/slog"
)

type netlifyidContextType int

const netlifyidContextKey netlifyidContextType = iota

func addJWTToRequest(id *JWT, r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), netlifyidContextKey, id)
	l := slog.
		FromContext(ctx).
		With("user.email", id.User.Email)

	return r.WithContext(slog.NewContext(ctx, l))
}

func FromContext(ctx context.Context) *JWT {
	val, _ := ctx.Value(netlifyidContextKey).(*JWT)
	return val
}
