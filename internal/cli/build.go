package cli

import (
	"fmt"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the registry manifest",
	Long: `Scans the registry directory for markdown files with regis3 frontmatter
and builds a manifest.json file in the .build directory.

The manifest is used for fast lookups and dependency resolution.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuild()
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func runBuild() error {
	debugf("Building manifest from: %s", getRegistryPath())

	result, err := registry.BuildRegistry(getRegistryPath())
	if err != nil {
		writer.Error(fmt.Sprintf("Build failed: %s", err.Error()))
		return err
	}

	itemCount := len(result.Manifest.Items)
	manifestPath := fmt.Sprintf("%s/.build/manifest.json", getRegistryPath())

	// Create response
	resp := output.NewResponseBuilder("build").
		WithSuccess(true).
		WithData(output.BuildData{
			ItemCount:    itemCount,
			ManifestPath: manifestPath,
			Duration:     result.Duration.String(),
		})

	// Add info about items found
	if itemCount > 0 {
		resp.WithInfo("Found %d items", itemCount)
	}

	// Add any errors from scanning
	for _, scanErr := range result.ScanErrors {
		resp.WithWarning("%s: %s", scanErr.Path, scanErr.Message)
	}

	writer.Write(resp.Build())
	return nil
}
