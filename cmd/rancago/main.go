package main

import (
	"os"

	"github.com/rancago/framework/internal/adapters/driving/cli"
)

func main() {
	os.Exit(cli.RunCLI())
}
