package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/okto-digital/regis3/internal/registry"
)

// itemOption represents a selectable item in the picker.
type itemOption struct {
	ref   string // type:name
	label string // display label
}

// pickItemsToAdd shows an interactive picker for selecting items to add.
func pickItemsToAdd(manifest *registry.Manifest) ([]string, error) {
	if len(manifest.Items) == 0 {
		return nil, fmt.Errorf("no items found in registry")
	}

	// Group items by type
	grouped := groupItemsByType(manifest)

	// Build options with visual grouping
	options := buildGroupedOptions(grouped)

	if len(options) == 0 {
		return nil, fmt.Errorf("no items available")
	}

	// Create huh select options (skip group headers, they're just for visual reference)
	var huhOptions []huh.Option[string]
	for _, opt := range options {
		if opt.ref != "" {
			huhOptions = append(huhOptions, huh.NewOption(opt.label, opt.ref))
		}
	}

	// Create multi-select
	var selected []string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select items to add").
				Description("Use arrow keys to navigate, space to select, enter to confirm").
				Options(huhOptions...).
				Value(&selected).
				Filterable(true).
				Limit(20),
		),
	)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Filter out empty selections (group headers)
	var result []string
	for _, s := range selected {
		if s != "" {
			result = append(result, s)
		}
	}

	return result, nil
}

// groupItemsByType groups manifest items by their type.
func groupItemsByType(manifest *registry.Manifest) map[string][]*registry.Item {
	grouped := make(map[string][]*registry.Item)

	for _, item := range manifest.Items {
		grouped[item.Type] = append(grouped[item.Type], item)
	}

	// Sort items within each group by name
	for itemType := range grouped {
		sort.Slice(grouped[itemType], func(i, j int) bool {
			return grouped[itemType][i].Name < grouped[itemType][j].Name
		})
	}

	return grouped
}

// buildGroupedOptions creates a flat list of options with group headers.
func buildGroupedOptions(grouped map[string][]*registry.Item) []itemOption {
	var options []itemOption

	// Define type order for consistent display
	typeOrder := []string{
		"skill", "subagent", "command", "doc", "prompt",
		"philosophy", "project", "ruleset",
		"mcp", "script", "hook", "stack",
	}

	// Add items in type order
	for _, itemType := range typeOrder {
		items, ok := grouped[itemType]
		if !ok || len(items) == 0 {
			continue
		}

		// Add group header
		header := fmt.Sprintf("── %s ──", strings.Title(itemType)+"s")
		options = append(options, itemOption{ref: "", label: header})

		// Add items
		for _, item := range items {
			ref := fmt.Sprintf("%s:%s", item.Type, item.Name)
			label := fmt.Sprintf("  %s", ref)
			if item.Desc != "" {
				// Truncate description if too long
				desc := item.Desc
				if len(desc) > 50 {
					desc = desc[:47] + "..."
				}
				label = fmt.Sprintf("  %-30s  %s", ref, desc)
			}
			options = append(options, itemOption{ref: ref, label: label})
		}
	}

	return options
}
