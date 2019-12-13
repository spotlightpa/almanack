package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/spotlightpa/almanack/pkg/worker"
)

func main() {
	exitcode.Exit(worker.CLI(os.Args[1:]))
}
