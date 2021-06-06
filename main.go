package main

import (
	"os"

	"github.com/janderland/tq/internal"
)

func main() {
	if err := internal.Run(); err != nil {
		os.Exit(1)
	}
}
