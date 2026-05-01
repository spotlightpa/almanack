package db

import (
	"flag"
)

// AddFlags adds an option to the specified FlagSet that creates and tests a db.Handle
func AddFlags(fl *flag.FlagSet, name, usage string) *Handle {
	h := new(Handle)
	fl.Func(name, usage, func(dbURL string) error {
		p, err := Open(dbURL)
		if p != nil {
			h.p = p
		}
		return err
	})
	return h
}
