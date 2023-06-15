package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	NullTime = pgtype.Timestamptz{}
	NullText = pgtype.Text{}
)
