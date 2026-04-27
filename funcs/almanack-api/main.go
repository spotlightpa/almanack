package main

import (
	"os"

	"github.com/spotlightpa/almanack/internal/almapp"
)

func main() {
	if err := almapp.CLI(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
