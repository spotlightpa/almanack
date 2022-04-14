package db

import (
	"reflect"

	"github.com/jackc/pgtype"
)

var (
	NullTime  = pgtype.Timestamptz{Status: pgtype.Null}
	NullJSONB = pgtype.JSONB{Status: pgtype.Null}
	NullText  = pgtype.Text{Status: pgtype.Null}
)

func IsPresent(s any) bool {
	v := reflect.ValueOf(s)
	status := v.FieldByName("Status").Interface().(pgtype.Status)
	return status == pgtype.Present
}

func IsNull(s any) bool {
	return !IsPresent(s)
}
