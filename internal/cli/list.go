package cli

import (
	"fmt"
	"sort"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var (
	listTypeFlag string
	listTagFlag  string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List items in the registry",
	Long: `Lists all items in the registry, optionally filtered by type or tag.

Examples:
  regis3 list                  # List all items
  regis3 list --type skill     # List only skills
  regis3 list --tag frontend   # List items with 'frontend' tag`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList()
	},
}

func init() {
	listCmd.Flags().StringVarP(&listTypeFlag, "type", "t", "", "Filter by type")
	listCmd.Flags().StringVar(&listTagFlag, "tag", "", "Filter by tag")
	rootCmd.AddCommand(listCmd)
}

func runList() error {
	debugf("Listing items from: %s", getRegistryPath())

	manifest, err := registry.LoadManifestFromRegistry(getRegistryPath())
	if err != nil {
		// Try building first
		debugf("Manifest not found, building...")
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

	// Convert map to slice and filter
	var items []*registry.Item
	for _, item := range manifest.Items {
		if listTypeFlag != "" && item.Type != listTypeFlag {
			continue
		}
		if listTagFlag != "" && !hasTag(item.Tags, listTagFlag) {
			continue
		}
		items = append(items, item)
	}

	// Sort by type then name
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type < items[j].Type
		}
		return items[i].Name < items[j].Name
	})

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

	resp := output.NewResponseBuilder("list").
		WithSuccess(true).
		WithData(output.ListData{
			Items:      listItems,
			TotalCount: len(manifest.Items),
			Filtered:   len(items) != len(manifest.Items),
		})

	if len(items) == 0 {
		if listTypeFlag != "" || listTagFlag != "" {
			resp.WithInfo("No items match the filter")
		} else {
			resp.WithInfo("Registry is empty")
		}
	}

	writer.Write(resp.Build())
	return nil
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
