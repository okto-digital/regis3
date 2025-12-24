// Package main is the entry point for the regis3 CLI.
package main

import (
	"fmt"
	"os"

	"github.com/okto-digital/regis3/internal/cli"
)

// Build-time variables (set via ldflags)
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
