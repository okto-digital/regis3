package cli

import (
	"fmt"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var reindexCmd = &cobra.Command{
	Use:     "reindex",
	Aliases: []string{"rebuild"},
	Short:   "Rebuild the registry manifest",
	Long: `Rebuilds the manifest.json file by scanning all registry files.

This is useful after manually moving or editing files in the registry.
It's equivalent to 'regis3 build' but named for clarity when used after
file operations.

Examples:
  regis3 reindex`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runReindex()
	},
}

func init() {
	rootCmd.AddCommand(reindexCmd)
}

func runReindex() error {
	debugf("Reindexing registry: %s", getRegistryPath())

	result, err := registry.BuildRegistry(getRegistryPath())
	if err != nil {
		writer.Error(fmt.Sprintf("Reindex failed: %s", err.Error()))
		return err
	}

	itemCount := len(result.Manifest.Items)
	manifestPath := fmt.Sprintf("%s/.build/manifest.json", getRegistryPath())

	resp := output.NewResponseBuilder("reindex").
		WithSuccess(true).
		WithData(output.BuildData{
			ItemCount:    itemCount,
			ManifestPath: manifestPath,
			Duration:     result.Duration.String(),
		}).
		WithInfo("Indexed %d items", itemCount)

	for _, scanErr := range result.ScanErrors {
		resp.WithWarning("%s: %s", scanErr.Path, scanErr.Message)
	}

	writer.Write(resp.Build())
	return nil
}
