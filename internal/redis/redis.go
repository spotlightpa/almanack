package redis

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/resperr"
	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
)

var ErrNil = redis.ErrNil

type Dialer = func() (redis.Conn, error)

type Logger interface {
	Printf(format string, v ...interface{})
}

type Store struct {
	rp *redis.Pool
	l  Logger
}

// New creates a Store and pings it.
func New(d Dialer, l Logger) (*Store, error) {
	c := Store{
		rp: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 4 * time.Minute,
			Dial:        d,
		},
		l: l,
	}

	if err := c.Ping(); err != nil {
		c.printf("connecting to Redis: %v", err)
		return nil, err
	}
	return &c, nil
}

func wrapErr(err error) error {
	if err == redis.ErrNil {
		return resperr.WithStatusCode(err, http.StatusNotFound)
	}
	return err
}

func (rs *Store) printf(format string, v ...interface{}) {
	if rs.l != nil {
		rs.l.Printf(format, v...)
	}
}

// Ping Redis
func (rs *Store) Ping() (err error) {
	rs.printf("do Redis PING")
	conn := rs.rp.Get()
	defer errutil.Defer(&err, conn.Close)

	_, err = conn.Do("PING")
	return wrapErr(err)
}

// Get calls GET in Redis and converts values from JSON bytes
func (rs *Store) Get(key string, getv interface{}) (err error) {
	rs.printf("do Redis GET %q", key)
	conn := rs.rp.Get()
	defer errutil.Defer(&err, conn.Close)

	getb, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return wrapErr(err)
	}
	return json.Unmarshal(getb, getv)
}

// Set converts values to JSON bytes and calls SET in Redis
func (rs *Store) Set(key string, setv interface{}) (err error) {
	rs.printf("do Redis SET %q", key)
	conn := rs.rp.Get()
	defer errutil.Defer(&err, conn.Close)

	setb, err := json.Marshal(setv)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, setb)
	return wrapErr(err)
}

// GetSet converts values to JSON bytes and calls GETSET in Redis
func (rs *Store) GetSet(key string, getv, setv interface{}) (err error) {
	rs.printf("do Redis GETSET %q", key)
	conn := rs.rp.Get()
	defer errutil.Defer(&err, conn.Close)

	setb, err := json.Marshal(setv)
	if err != nil {
		return err
	}
	getb, err := redis.Bytes(conn.Do("GETSET", key, setb))
	if err != nil {
		return wrapErr(err)
	}
	return json.Unmarshal(getb, getv)
}

func (rs *Store) GetLock(key string) (unlock func(), err error) {
	rs.printf("get Redis lock %q", key)
	lock := redsync.
		New([]redsync.Pool{rs.rp}).
		NewMutex(key)
	if err := lock.Lock(); err != nil {
		return nil, err
	}
	return func() {
		rs.printf("return Redis lock %q", key)
		lock.Unlock()
	}, nil
}
