package db

import "github.com/jackc/pgtype"

var (
	NullTime  = pgtype.Timestamptz{Status: pgtype.Null}
	NullJSONB = pgtype.JSONB{Status: pgtype.Null}
	NullText  = pgtype.Text{Status: pgtype.Null}
)
