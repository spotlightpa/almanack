package timeutil

import "time"

func ToEST(t time.Time) time.Time {
	newYork, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic("timezone db not available")
	}
	return t.In(newYork)
}
