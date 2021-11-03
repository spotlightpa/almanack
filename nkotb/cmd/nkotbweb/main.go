package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/nkotb/nkotbweb"
)

func main() {
	exitcode.Exit(nkotbweb.CLI(os.Args[1:]))
}
