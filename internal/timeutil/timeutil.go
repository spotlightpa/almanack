package timeutil

import (
	"time"

	"github.com/jackc/pgtype"
)

func ToEST(t time.Time) time.Time {
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic("timezone db not available")
	}
	return t.In(newYork)
}

func GetTime(m map[string]interface{}, key string) (t time.Time, ok bool) {
	return ToTime(m[key])
}

func ToTime(v interface{}) (t time.Time, ok bool) {
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
	if old.Status != new.Status {
		return false
	}
	if old.Status != pgtype.Present || new.Status != pgtype.Present {
		return true
	}
	// If the dates are more than 290 years apart
	// (e.g. because one is zero time)
	// the duration is the max that fits into an int64, but
	// math.MinInt64 can't be converted to a positive number.
	// To workaround this, just pre-sort them
	// so the diff will always be zero or positive.
	older, newer := old.Time, new.Time
	if newer.Before(older) {
		older, newer = new.Time, old.Time
	}
	diff := newer.Sub(older)
	return diff < timeWindow
}
