package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/nkotb/pkg/clis"
)

func main() {
	exitcode.Exit(clis.GDocs(os.Args[1:]))
}
