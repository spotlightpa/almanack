package timeutil

import "time"

func ToEST(t time.Time) time.Time {
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic("timezone db not available")
	}
	return t.In(newYork)
}

func GetTime(m map[string]interface{}, key string) (t time.Time, ok bool) {
	if t, _ = m[key].(time.Time); !t.IsZero() {
		ok = true
		return
	}
	s, _ := m[key].(string)
	if s == "" {
		return
	}
	ok = t.UnmarshalText([]byte(s)) == nil
	return
}
