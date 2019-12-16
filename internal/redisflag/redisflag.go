package redisflag

import (
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

type Dialer = func() (redis.Conn, error)

func Value(dialer *Dialer) flag.Getter {
	return &getter{dialer}
}

type getter struct {
	dialer *Dialer
}

func (g *getter) Get() interface{} { return g.dialer }

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
	db := 0
	if path := strings.TrimPrefix(u.Path, "/"); path != "" {
		if db, err = strconv.Atoi(path); err != nil {
			return err
		}
	}
	*g.dialer = func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", u.Host)
		if err != nil {
			return nil, err
		}
		if password != "" {
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
		}
		if db != 0 {
			if _, err := c.Do("SELECT", db); err != nil {
				c.Close()
				return nil, err
			}
		}
		return c, nil
	}
	return nil
}
