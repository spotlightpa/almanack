package errutil

import (
	"errors"

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
