package filestore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/carlmjohnson/resperr"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type Loc struct {
	dir string
	l   Logger
}

func New(base, ns string, l Logger) Loc {
	if base == "" {
		dir, err := os.UserCacheDir()
		if err != nil {
			l.Printf("warning: could not open user cache directory: %v", err)
		}
		base = dir
	}
	return Loc{filepath.Join(base, ns), l}
}

func (loc Loc) ensure() error {
	if err := os.MkdirAll(loc.dir, os.ModePerm); err != nil {
		return fmt.Errorf("problem with cache folder: %w", err)
	}
	return nil
}

func (loc Loc) name(key string) string {
	key = filepath.Clean(key)
	key = strings.ReplaceAll(key, string(filepath.Separator), "@")
	return filepath.Join(loc.dir, fmt.Sprintf("%s.json", key))
}

func (loc Loc) printf(format string, v ...interface{}) {
	if loc.l != nil {
		loc.l.Printf(format, v...)
	}
}

func (loc Loc) Get(key string, v interface{}) error {
	loc.printf("do file GET %q", loc.name(key))

	err := loc.ensure()
	if err != nil {
		return fmt.Errorf("could not read cache data: %w", err)
	}
	data, err := os.ReadFile(loc.name(key))
	if os.IsNotExist(err) {
		return resperr.WithStatusCode(err, http.StatusNotFound)
	} else if err != nil {
		return fmt.Errorf("could not read cache data: %w", err)
	}
	if err = json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("could not read cache data: %w", err)
	}
	return nil
}

func (loc Loc) Set(key string, v interface{}) error {
	loc.printf("do file SET %q", loc.name(key))

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("could not write cache data: %w", err)
	}
	if err = loc.ensure(); err != nil {
		return fmt.Errorf("could not write cache data: %w", err)
	}
	if err = os.WriteFile(loc.name(key), data, 0644); err != nil {
		return fmt.Errorf("could not write cache data: %w", err)
	}
	return nil
}

func (loc Loc) GetSet(key string, getv, setv interface{}) (err error) {
	loc.printf("do file GETSET %q", loc.name(key))

	geterr := loc.Get(key, getv)
	if err = loc.Set(key, setv); err != nil {
		return err
	}
	return geterr
}

func (loc Loc) GetLock(key string) (unlock func(), err error) {
	// TODO: actual locking
	loc.printf("noop getlock %q", key)
	return func() {
		loc.printf("noop unlock %q", key)
	}, nil
}
