package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/nkotb/pkg/clis"
)

func main() {
	exitcode.Exit(clis.NKOTB(os.Args[1:]))
}
