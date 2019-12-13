package redisflag

import (
	"flag"
	"fmt"
	"net/url"
	"time"

	"github.com/gomodule/redigo/redis"
)

func Value(pool *redis.Pool) flag.Getter {
	return &getter{pool}
}

type getter struct {
	pool *redis.Pool
}

func (g *getter) Get() interface{} { return g.pool }

func (g *getter) String() string { return "Redis URL" }

func (g *getter) Set(connURL string) error {
	u, err := url.Parse(connURL)
	if err != nil {
		return err
	}
	if u.Scheme != "redis" {
		return fmt.Errorf("invalid redis URL scheme: %q", u.Scheme)
	}
	password, _ := u.User.Password()

	dialer := func() (redis.Conn, error) { return redis.Dial("tcp", u.Host) }
	if password != "" {
		dialer = func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", u.Host)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		}
	}
	*g.pool = redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial:        dialer,
	}
	return nil
}
