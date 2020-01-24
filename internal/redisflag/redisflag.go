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

// Parse creates a Dialer by parsing the connection URL.
//
// URLs should have the format redis://user:password@host:port/db
// where db is an integer and username is ignored.
func Parse(connURL string) (Dialer, error) {
	u, err := url.Parse(connURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "redis" {
		return nil, fmt.Errorf("invalid redis URL scheme: %q", u.Scheme)
	}
	password, _ := u.User.Password()
	db := 0
	if path := strings.TrimPrefix(u.Path, "/"); path != "" {
		if db, err = strconv.Atoi(path); err != nil {
			return nil, err
		}
	}
	return func() (redis.Conn, error) {
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
	}, nil
}

type getter struct {
	dialer Dialer
}

func (g *getter) Get() interface{} { return g.dialer }

func (g *getter) String() string { return "Redis URL" }

func (g *getter) Set(connURL string) error {
	var err error
	g.dialer, err = Parse(connURL)
	return err
}

// Var adds an option to the specified FlagSet (or flag.CommandLine if nil)
// that creates a Redis dialer for the specified URL.
// URLs should have the format redis://user:password@host:port/db
// where db is an integer and username is ignored.
// Use the callback after parsing options to retrieve the dialer.
func Var(fl *flag.FlagSet, name, usage string) func() Dialer {
	if fl == nil {
		fl = flag.CommandLine
	}
	var g getter
	fl.Var(&g, name, usage)
	return func() Dialer {
		return g.dialer
	}
}
