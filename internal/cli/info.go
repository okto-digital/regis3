package cli

import (
	"fmt"
	"strings"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <type:name>",
	Short: "Show item details",
	Long: `Shows detailed information about a registry item.

Examples:
  regis3 info skill:git-conventions
  regis3 info subagent:code-reviewer`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInfo(args[0])
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func runInfo(ref string) error {
	// Parse type:name reference
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		writer.Error("Invalid reference format. Use 'type:name' (e.g., skill:git-conventions)")
		return fmt.Errorf("invalid reference format")
	}

	itemType, itemName := parts[0], parts[1]
	debugf("Looking up: %s:%s", itemType, itemName)

	manifest, err := registry.LoadManifestFromRegistry(getRegistryPath())
	if err != nil {
		// Try building first
		_, buildErr := registry.BuildRegistry(getRegistryPath())
		if buildErr != nil {
			writer.Error(fmt.Sprintf("Failed to load registry: %s", err.Error()))
			return err
		}
		manifest, err = registry.LoadManifestFromRegistry(getRegistryPath())
		if err != nil {
			writer.Error(fmt.Sprintf("Failed to load manifest: %s", err.Error()))
			return err
		}
	}

	// Find the item
	fullName := fmt.Sprintf("%s:%s", itemType, itemName)
	item, ok := manifest.Items[fullName]
	if !ok {
		writer.Error(fmt.Sprintf("Item '%s' not found in registry", fullName))
		return fmt.Errorf("item not found")
	}

	// Build info data
	infoData := output.InfoData{
		Type:         item.Type,
		Name:         item.Name,
		Desc:         item.Desc,
		Path:         item.Source,
		Tags:         item.Tags,
		Dependencies: item.Deps,
		Files:        item.Files,
	}

	resp := output.NewResponseBuilder("info").
		WithSuccess(true).
		WithData(infoData)

	writer.Write(resp.Build())
	return nil
}
