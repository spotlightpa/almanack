package netlifyid

import (
	"context"
	"net/http"

	"github.com/spotlightpa/almanack/pkg/almlog"
)

type netlifyidContextType int

const netlifyidContextKey netlifyidContextType = iota

func addJWTToRequest(id *JWT, r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), netlifyidContextKey, id)
	l := almlog.FromContext(ctx).
		With("user.email", id.User.Email)

	return r.WithContext(almlog.NewContext(ctx, l))
}

func FromContext(ctx context.Context) *JWT {
	val, _ := ctx.Value(netlifyidContextKey).(*JWT)
	return val
}
