package netlifyid

import (
	"context"
	"net/http"
)

type netlifyidContextType int

const netlifyidContextKey netlifyidContextType = iota

func addJWTToRequest(id *JWT, r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), netlifyidContextKey, id)
	return r.WithContext(ctx)
}

func FromContext(ctx context.Context) *JWT {
	val, _ := ctx.Value(netlifyidContextKey).(*JWT)
	return val
}
