package pgxutil

import (
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	NullTime = pgtype.Timestamptz{}
	NullText = pgtype.Text{}
)

// NilSliceToEmpty turns nil slices into empty slices
// for cases where pgx treats nil as NULL.
func NilSliceToEmpty[S ~[]Elem, Elem any](s S) S {
	if s == nil {
		return make(S, 0)
	}
	return s
}
