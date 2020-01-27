package errutil

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/spotlightpa/almanack/internal/redis"
)

type Type int8

const (
	NotFound Type = iota
)

var names = []string{
	NotFound: "Not Found",
}

func (t Type) Error() string {
	return names[int(t)]
}

func Is(err error, t Type) bool {
	switch t {
	case NotFound:
		if errors.Is(err, redis.ErrNil) {
			return true
		}
	}
	if errors.Is(err, t) {
		return true
	}
	return false
}

var codes = []int{
	NotFound: http.StatusNotFound,
}

func (t Type) As(v interface{}) bool {
	if _, ok := v.(*Response); ok {
		if n := int(t); n < 0 && n >= len(codes) {
			return false
		}
		r := Response{
			StatusCode: codes[int(t)],
			Message:    t.Error(),
			Log:        t.Error(),
		}
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(r))
		return true
	}
	return false
}
