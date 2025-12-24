// Package main is the entry point for the regis3 CLI.
package main

import (
	"fmt"
	"os"
)

// Build-time variables (set via ldflags)
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// TODO: Replace with Cobra root command in Phase 6
	fmt.Printf("regis3 %s (built %s)\n", version, buildTime)
	fmt.Println("Configuration registry for LLM assistants")
	fmt.Println()
	fmt.Println("Commands will be implemented in Phase 6.")
	fmt.Println("Run 'make test' to verify the foundations work.")
	os.Exit(0)
}
