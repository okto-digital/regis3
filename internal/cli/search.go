package cli

import (
	"fmt"
	"strings"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search items in the registry",
	Long: `Searches items by name, description, and tags.

Examples:
  regis3 search git           # Find items containing 'git'
  regis3 search "clean code"  # Find items containing 'clean code'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearch(args[0])
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

func runSearch(query string) error {
	debugf("Searching for: %s", query)

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

	// Search items
	items := searchItems(manifest.Items, query)

	// Build list data
	listItems := make([]output.ListItem, len(items))
	for i, item := range items {
		listItems[i] = output.ListItem{
			Type: item.Type,
			Name: item.Name,
			Desc: item.Desc,
			Tags: item.Tags,
		}
	}

	resp := output.NewResponseBuilder("search").
		WithSuccess(true).
		WithData(output.ListData{
			Items:      listItems,
			TotalCount: len(manifest.Items),
			Filtered:   true,
		})

	if len(items) == 0 {
		resp.WithInfo("No items match '%s'", query)
	} else {
		resp.WithInfo("Found %d items matching '%s'", len(items), query)
	}

	writer.Write(resp.Build())
	return nil
}

func searchItems(items map[string]*registry.Item, query string) []*registry.Item {
	query = strings.ToLower(query)
	var matches []*registry.Item

	for _, item := range items {
		// Search in name
		if strings.Contains(strings.ToLower(item.Name), query) {
			matches = append(matches, item)
			continue
		}

		// Search in description
		if strings.Contains(strings.ToLower(item.Desc), query) {
			matches = append(matches, item)
			continue
		}

		// Search in tags
		for _, tag := range item.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				matches = append(matches, item)
				break
			}
		}
	}

	return matches
}
