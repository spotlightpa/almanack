package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/almanack/pkg/almanack"
)

func main() {
	exitcode.Exit(almanack.CLI(os.Args[1:]))
}
