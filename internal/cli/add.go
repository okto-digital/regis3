package cli

import (
	"fmt"
	"strings"

	"github.com/okto-digital/regis3/internal/installer"
	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var (
	addDryRun bool
	addForce  bool
	addTarget string
)

var addCmd = &cobra.Command{
	Use:   "add <type:name> [type:name...]",
	Short: "Install items to the project",
	Long: `Installs one or more items from the registry to the current project.

Dependencies are automatically resolved and installed in the correct order.

Examples:
  regis3 add skill:git-conventions
  regis3 add skill:git-conventions skill:clean-code
  regis3 add stack:vue-fullstack`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAdd(args)
	},
}

func init() {
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "Preview what would be installed")
	addCmd.Flags().BoolVarP(&addForce, "force", "F", false, "Force reinstall even if already installed")
	addCmd.Flags().StringVar(&addTarget, "target", "", "Target (default: from config)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(refs []string) error {
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
	targetName := addTarget
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
		// Try to load from targets directory
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
	inst.DryRun = addDryRun
	inst.Force = addForce

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

	resp := output.NewResponseBuilder("add").
		WithData(output.InstallData{
			Installed: installed,
			Skipped:   result.Skipped,
			Target:    targetName,
			DryRun:    addDryRun,
		})

	if len(result.Errors) > 0 {
		resp.WithSuccess(false)
		for _, e := range result.Errors {
			resp.WithError(e.ItemID, e.Message)
		}
	} else {
		resp.WithSuccess(true)
		if addDryRun {
			resp.WithInfo("Would install %d items (dry run)", len(installed))
		} else if len(installed) > 0 {
			resp.WithInfo("Installed %d items", len(installed))
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
