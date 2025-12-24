package cli

import (
	"fmt"
	"strings"

	"github.com/okto-digital/regis3/internal/installer"
	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

// Project command flags
var (
	projectAddDryRun    bool
	projectAddForce     bool
	projectAddTarget    string
	projectRemoveDryRun bool
	projectRemoveTarget string
	projectStatusTarget string
)

// projectCmd is the parent command for project operations
var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage items in the current project",
	Long: `Commands for adding, removing, and viewing items installed in the current project.

These commands operate on the current working directory, installing items from
your registry into the project's .claude/ directory (for Claude Code target).

Examples:
  regis3 project add skill:git-conventions
  regis3 project remove skill:testing
  regis3 project status`,
}

// projectAddCmd installs items to the current project
var projectAddCmd = &cobra.Command{
	Use:   "add <type:name> [type:name...]",
	Short: "Add items to the current project",
	Long: `Installs one or more items from the registry to the current project.

Dependencies are automatically resolved and installed in the correct order.
Items are installed to the .claude/ directory (for Claude Code target).

Examples:
  regis3 project add skill:git-conventions
  regis3 project add skill:git-conventions skill:clean-code
  regis3 project add stack:vue-fullstack`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no args provided, show interactive picker
		if len(args) == 0 {
			manifest, err := registry.LoadManifestFromRegistry(getRegistryPath())
			if err != nil {
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

			selected, err := pickItemsToAdd(manifest)
			if err != nil {
				writer.Error(fmt.Sprintf("Selection cancelled: %s", err.Error()))
				return err
			}

			if len(selected) == 0 {
				writer.Info("No items selected")
				return nil
			}

			args = selected
		}
		return runProjectAdd(args)
	},
}

// projectRemoveCmd removes items from the current project
var projectRemoveCmd = &cobra.Command{
	Use:     "remove <type:name> [type:name...]",
	Aliases: []string{"rm"},
	Short:   "Remove items from the current project",
	Long: `Removes one or more installed items from the current project.

Examples:
  regis3 project remove skill:git-conventions
  regis3 project rm skill:git-conventions skill:clean-code`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("missing item reference\n\nUsage: regis3 project remove <type:name> [type:name...]\n\nExample: regis3 project remove skill:git-conventions")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectRemove(args)
	},
}

// projectStatusCmd shows installed items in the current project
var projectStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show items installed in the current project",
	Long: `Shows all items currently installed in the current project.

Examples:
  regis3 project status
  regis3 project status --target claude`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runProjectStatus()
	},
}

func init() {
	// Add flags
	projectAddCmd.Flags().BoolVar(&projectAddDryRun, "dry-run", false, "Preview what would be installed")
	projectAddCmd.Flags().BoolVarP(&projectAddForce, "force", "F", false, "Force reinstall even if already installed")
	projectAddCmd.Flags().StringVar(&projectAddTarget, "target", "", "Target (default: from config)")

	projectRemoveCmd.Flags().BoolVar(&projectRemoveDryRun, "dry-run", false, "Preview what would be removed")
	projectRemoveCmd.Flags().StringVar(&projectRemoveTarget, "target", "", "Target (default: from config)")

	projectStatusCmd.Flags().StringVar(&projectStatusTarget, "target", "", "Target (default: from config)")

	// Add subcommands to project
	projectCmd.AddCommand(projectAddCmd)
	projectCmd.AddCommand(projectRemoveCmd)
	projectCmd.AddCommand(projectStatusCmd)

	// Add project to root
	rootCmd.AddCommand(projectCmd)
}

func runProjectAdd(refs []string) error {
	// Validate references
	for _, ref := range refs {
		if !strings.Contains(ref, ":") {
			writer.Error(fmt.Sprintf("Invalid reference '%s' - use format 'type:name'", ref))
			return fmt.Errorf("invalid reference: %s", ref)
		}
	}

	// Load manifest
	manifest, err := registry.LoadManifestFromRegistry(getRegistryPath())
	if err != nil {
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

	// Get target
	targetName := projectAddTarget
	if targetName == "" && cfg != nil {
		targetName = cfg.DefaultTarget
	}
	if targetName == "" {
		targetName = "claude"
	}

	// Get target config
	var target *installer.Target
	if targetName == "claude" {
		target = installer.DefaultClaudeTarget()
	} else {
		target, err = installer.LoadTargetByName("targets", targetName)
		if err != nil {
			writer.Error(fmt.Sprintf("Target not found: %s", err.Error()))
			return err
		}
	}

	// Create installer
	inst, err := installer.NewInstaller(".", getRegistryPath(), target)
	if err != nil {
		writer.Error(fmt.Sprintf("Installer error: %s", err.Error()))
		return err
	}
	inst.DryRun = projectAddDryRun
	inst.Force = projectAddForce

	// Install items
	result, err := inst.Install(manifest, refs)
	if err != nil {
		writer.Error(fmt.Sprintf("Installation failed: %s", err.Error()))
		return err
	}

	// Build response
	var installed []output.InstalledItem
	for _, id := range result.Installed {
		parts := strings.SplitN(id, ":", 2)
		if len(parts) == 2 {
			installed = append(installed, output.InstalledItem{
				Type: parts[0],
				Name: parts[1],
			})
		}
	}
	for _, id := range result.Updated {
		parts := strings.SplitN(id, ":", 2)
		if len(parts) == 2 {
			installed = append(installed, output.InstalledItem{
				Type: parts[0],
				Name: parts[1],
			})
		}
	}

	resp := output.NewResponseBuilder("project add").
		WithData(output.InstallData{
			Installed: installed,
			Skipped:   result.Skipped,
			Target:    targetName,
			DryRun:    projectAddDryRun,
		})

	if len(result.Errors) > 0 {
		resp.WithSuccess(false)
		for _, e := range result.Errors {
			resp.WithError(e.ItemID, e.Message)
		}
	} else {
		resp.WithSuccess(true)
		if projectAddDryRun {
			resp.WithInfo("Would install %d items (dry run)", len(installed))
		} else if len(installed) > 0 {
			resp.WithInfo("Installed %d items to project", len(installed))
		}
		if len(result.Skipped) > 0 {
			resp.WithInfo("Skipped %d already installed", len(result.Skipped))
		}
		if len(result.MergedItems) > 0 {
			resp.WithInfo("Merged %d items into %s", len(result.MergedItems), target.MergeFile)
		}
	}

	writer.Write(resp.Build())

	if len(result.Errors) > 0 {
		return fmt.Errorf("installation failed")
	}
	return nil
}

func runProjectRemove(refs []string) error {
	// Get target
	targetName := projectRemoveTarget
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

	// Create installer
	inst, err := installer.NewInstaller(".", getRegistryPath(), target)
	if err != nil {
		writer.Error(fmt.Sprintf("Installer error: %s", err.Error()))
		return err
	}
	inst.DryRun = projectRemoveDryRun

	// Uninstall items
	result, err := inst.Uninstall(refs)
	if err != nil {
		writer.Error(fmt.Sprintf("Uninstall failed: %s", err.Error()))
		return err
	}

	// Build response
	var removed []output.InstalledItem
	for _, id := range result.Uninstalled {
		parts := strings.SplitN(id, ":", 2)
		if len(parts) == 2 {
			removed = append(removed, output.InstalledItem{
				Type: parts[0],
				Name: parts[1],
			})
		}
	}

	resp := output.NewResponseBuilder("project remove").
		WithData(output.RemoveData{
			Removed:  removed,
			NotFound: result.NotFound,
			DryRun:   projectRemoveDryRun,
		})

	if len(result.Errors) > 0 {
		resp.WithSuccess(false)
		for _, e := range result.Errors {
			resp.WithError(e.ItemID, e.Message)
		}
	} else {
		resp.WithSuccess(true)
		if projectRemoveDryRun {
			resp.WithInfo("Would remove %d items (dry run)", len(removed))
		} else if len(removed) > 0 {
			resp.WithInfo("Removed %d items from project", len(removed))
		}
		if len(result.NotFound) > 0 {
			resp.WithWarning("%d items not installed", len(result.NotFound))
		}
		if len(result.Skipped) > 0 {
			resp.WithWarning("Skipped %d merged items (edit %s manually)", len(result.Skipped), target.MergeFile)
		}
	}

	writer.Write(resp.Build())

	if len(result.Errors) > 0 {
		return fmt.Errorf("removal failed")
	}
	return nil
}

func runProjectStatus() error {
	// Get target
	targetName := projectStatusTarget
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

	resp := output.NewResponseBuilder("project status").
		WithSuccess(true).
		WithData(output.StatusData{
			Items:  items,
			Target: targetName,
		})

	if len(items) == 0 {
		resp.WithInfo("No items installed in this project")
	} else {
		resp.WithInfo("%d items installed in this project", len(items))

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
