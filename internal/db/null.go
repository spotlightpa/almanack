package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	NullTime = pgtype.Timestamptz{}
	NullText = pgtype.Text{}
)

func Array[T any](elems ...T) pgtype.Array[T] {
	return pgtype.Array[T]{
		Elements: elems,
		Dims: []pgtype.ArrayDimension{
			{Length: int32(len(elems))},
		},
		Valid: true,
	}
}
