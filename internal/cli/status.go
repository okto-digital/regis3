package cli

import (
	"fmt"

	"github.com/okto-digital/regis3/internal/installer"
	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var statusTarget string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed items",
	Long: `Shows all items installed in the current project.

Examples:
  regis3 status
  regis3 status --target claude`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusTarget, "target", "", "Target (default: from config)")
	rootCmd.AddCommand(statusCmd)
}

func runStatus() error {
	// Get target
	targetName := statusTarget
	if targetName == "" && cfg != nil {
		targetName = cfg.DefaultTarget
	}
	if targetName == "" {
		targetName = "claude"
	}

	// Get target config
	var target *installer.Target
	var err error
	if targetName == "claude" {
		target = installer.DefaultClaudeTarget()
	} else {
		target, err = installer.LoadTargetByName("targets", targetName)
		if err != nil {
			writer.Error(fmt.Sprintf("Target not found: %s", err.Error()))
			return err
		}
	}

	// Load manifest (needed for status check)
	manifest, err := registry.LoadManifestFromRegistry(getRegistryPath())
	if err != nil {
		// If no manifest, just show tracker status
		manifest = &registry.Manifest{Items: make(map[string]*registry.Item)}
	}

	// Create installer to access status
	inst, err := installer.NewInstaller(".", getRegistryPath(), target)
	if err != nil {
		writer.Error(fmt.Sprintf("Error: %s", err.Error()))
		return err
	}

	// Get status
	status := inst.Status(manifest)

	// Build status items (only installed ones)
	var items []output.StatusItem
	for _, s := range status.Items {
		if s.Installed {
			installedAt := ""
			if s.InstalledAt != nil {
				if t, ok := s.InstalledAt.(string); ok {
					installedAt = t
				}
			}
			items = append(items, output.StatusItem{
				Type:        s.Type,
				Name:        s.Name,
				InstalledAt: installedAt,
				DestPath:    s.Path,
				NeedsUpdate: s.NeedsUpdate,
			})
		}
	}

	resp := output.NewResponseBuilder("status").
		WithSuccess(true).
		WithData(output.StatusData{
			Items:  items,
			Target: targetName,
		})

	if len(items) == 0 {
		resp.WithInfo("No items installed")
	} else {
		resp.WithInfo("%d items installed", len(items))

		// Check for updates
		updateCount := 0
		for _, item := range items {
			if item.NeedsUpdate {
				updateCount++
			}
		}
		if updateCount > 0 {
			resp.WithWarning("%d items have updates available", updateCount)
		}
	}

	writer.Write(resp.Build())
	return nil
}
