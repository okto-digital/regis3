package cli

import (
	"fmt"
	"strings"

	"github.com/okto-digital/regis3/internal/installer"
	"github.com/okto-digital/regis3/internal/output"
	"github.com/spf13/cobra"
)

var (
	removeDryRun bool
	removeTarget string
)

var removeCmd = &cobra.Command{
	Use:     "remove <type:name> [type:name...]",
	Aliases: []string{"rm", "uninstall"},
	Short:   "Remove installed items",
	Long: `Removes one or more installed items from the current project.

Examples:
  regis3 remove skill:git-conventions
  regis3 rm skill:git-conventions skill:clean-code`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRemove(args)
	},
}

func init() {
	removeCmd.Flags().BoolVar(&removeDryRun, "dry-run", false, "Preview what would be removed")
	removeCmd.Flags().StringVar(&removeTarget, "target", "", "Target (default: from config)")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(refs []string) error {
	// Get target
	targetName := removeTarget
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
	inst.DryRun = removeDryRun

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

	resp := output.NewResponseBuilder("remove").
		WithData(output.RemoveData{
			Removed:  removed,
			NotFound: result.NotFound,
			DryRun:   removeDryRun,
		})

	if len(result.Errors) > 0 {
		resp.WithSuccess(false)
		for _, e := range result.Errors {
			resp.WithError(e.ItemID, e.Message)
		}
	} else {
		resp.WithSuccess(true)
		if removeDryRun {
			resp.WithInfo("Would remove %d items (dry run)", len(removed))
		} else if len(removed) > 0 {
			resp.WithInfo("Removed %d items", len(removed))
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
