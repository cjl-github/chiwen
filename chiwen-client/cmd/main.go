package main

import (
	"os"

	"github.com/chiwen/client/cmd/app"
)

func main() {
	cmd := app.NewServerCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
