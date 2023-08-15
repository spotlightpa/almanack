package timex

import (
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/spotlightpa/almanack/internal/must"
)

var getNewYork = sync.OnceValue(func() *time.Location {
	return must.Get(time.LoadLocation("America/New_York"))
})

func ToEST(t time.Time) time.Time {
	return t.In(getNewYork())
}

func Unwrap(v any) (t time.Time, ok bool) {
	if t, _ = v.(time.Time); !t.IsZero() {
		ok = true
		return
	}
	s, _ := v.(string)
	if s == "" {
		return
	}
	ok = t.UnmarshalText([]byte(s)) == nil
	return
}

const timeWindow = 5 * time.Minute

func Equalish(old, new pgtype.Timestamptz) bool {
	if old.Valid != new.Valid {
		return false
	}
	if !old.Valid {
		return true
	}
	diff := old.Time.Sub(new.Time).Abs()
	return diff < timeWindow
}
