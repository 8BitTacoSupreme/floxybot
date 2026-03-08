package main

import (
	"fmt"
	"os"

	"github.com/8BitTacoSupreme/floxybot/internal/cli"
)

// Set by -ldflags at build time.
var version = "dev"

func main() {
	cli.SetVersion(version)
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
