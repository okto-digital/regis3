package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information (set via ldflags by goreleaser)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
	builtBy = "manual"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion() error {
	fmt.Printf("regis3 %s\n", version)
	fmt.Printf("  Built:    %s\n", date)
	fmt.Printf("  Commit:   %s\n", commit)
	fmt.Printf("  Go:       %s\n", runtime.Version())
	fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	if builtBy != "manual" {
		fmt.Printf("  Built by: %s\n", builtBy)
	}
	return nil
}
