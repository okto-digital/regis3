package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the registry from git remote",
	Long: `Updates the registry by pulling the latest changes from git.

This command runs 'git pull' in the registry directory to fetch
the latest items from the remote repository.

After pulling, it automatically rebuilds the manifest.

Examples:
  regis3 update`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate() error {
	registryPath := getRegistryPath()
	debugf("Updating registry: %s", registryPath)

	// Check if registry is a git repo
	gitDir := filepath.Join(registryPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		writer.Error("Registry is not a git repository")
		return fmt.Errorf("not a git repository")
	}

	// Run git pull
	cmd := exec.Command("git", "-C", registryPath, "pull", "--ff-only")
	gitOutput, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(gitOutput))

	if err != nil {
		writer.Error(fmt.Sprintf("Git pull failed: %s", outputStr))
		return err
	}

	// Check if there were updates
	alreadyUpToDate := strings.Contains(outputStr, "Already up to date")

	// Rebuild manifest
	result, err := registry.BuildRegistry(registryPath)
	if err != nil {
		writer.Error(fmt.Sprintf("Failed to rebuild manifest: %s", err.Error()))
		return err
	}

	itemCount := len(result.Manifest.Items)

	resp := output.NewResponseBuilder("update").
		WithSuccess(true).
		WithData(output.UpdateData{
			Updated:   !alreadyUpToDate,
			ItemCount: itemCount,
			GitOutput: outputStr,
		})

	if alreadyUpToDate {
		resp.WithInfo("Registry is already up to date (%d items)", itemCount)
	} else {
		resp.WithSuccess(true)
		resp.WithInfo("Registry updated (%d items)", itemCount)
	}

	writer.Write(resp.Build())
	return nil
}
