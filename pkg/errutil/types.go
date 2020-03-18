package errutil

import (
	"net/http"
	"reflect"
)

type Type int8

const (
	NotFound Type = iota
	Unauthorized
)

var names = []string{
	NotFound:     "Not Found",
	Unauthorized: "Unauthorized",
}

func (t Type) Error() string {
	return names[int(t)]
}

var codes = []int{
	NotFound:     http.StatusNotFound,
	Unauthorized: http.StatusForbidden,
}

func (t Type) As(v interface{}) bool {
	if _, ok := v.(*Response); ok {
		if n := int(t); n < 0 && n >= len(codes) {
			return false
		}
		r := Response{
			StatusCode: codes[int(t)],
			Message:    t.Error(),
			Cause:      t,
		}
		reflect.ValueOf(v).Elem().Set(reflect.ValueOf(r))
		return true
	}
	return false
}
