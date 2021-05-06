package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/nkotb/blocko"
)

func main() {
	exitcode.Exit(blocko.CLI(os.Args[1:]))
}
