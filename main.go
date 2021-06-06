package main

import (
	"os"

	"github.com/janderland/tq/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
