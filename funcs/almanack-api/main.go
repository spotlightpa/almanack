package main

import (
	"os"

	"github.com/spotlightpa/almanack/pkg/api"
)

func main() {
	if err := api.CLI(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
