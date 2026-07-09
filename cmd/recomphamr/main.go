package main

import (
	"os"

	"recomphamr2/internal/app"
)

var exit = os.Exit

func main() {
	exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
