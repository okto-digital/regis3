package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var orphansCmd = &cobra.Command{
	Use:   "orphans",
	Short: "Find unreferenced files in the registry",
	Long: `Finds markdown files in the registry that are not included in the manifest.

This can happen when:
- Files don't have valid regis3 frontmatter
- Files were manually moved without updating frontmatter
- Files are in excluded directories

Examples:
  regis3 orphans`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runOrphans()
	},
}

func init() {
	rootCmd.AddCommand(orphansCmd)
}

func runOrphans() error {
	registryPath := getRegistryPath()
	debugf("Scanning for orphans in: %s", registryPath)

	// Load manifest
	manifest, err := registry.LoadManifestFromRegistry(registryPath)
	if err != nil {
		// Try building first
		_, buildErr := registry.BuildRegistry(registryPath)
		if buildErr != nil {
			writer.Error("Failed to load registry")
			return err
		}
		manifest, err = registry.LoadManifestFromRegistry(registryPath)
		if err != nil {
			writer.Error("Failed to load manifest")
			return err
		}
	}

	// Build set of known files
	knownFiles := make(map[string]bool)
	for _, item := range manifest.Items {
		knownFiles[item.Source] = true
		for _, f := range item.Files {
			knownFiles[f] = true
		}
	}

	// Find all markdown files
	var orphans []output.OrphanFile
	err = filepath.Walk(registryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Skip directories
		if info.IsDir() {
			name := info.Name()
			// Skip hidden directories and build directory
			if strings.HasPrefix(name, ".") || name == "import" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only check markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(registryPath, path)
		if err != nil {
			return nil
		}

		// Check if it's in the manifest
		if !knownFiles[relPath] {
			orphans = append(orphans, output.OrphanFile{
				Path:   relPath,
				Size:   info.Size(),
				Reason: "not in manifest",
			})
		}

		return nil
	})

	if err != nil {
		writer.Error("Failed to scan registry")
		return err
	}

	resp := output.NewResponseBuilder("orphans").
		WithSuccess(true).
		WithData(output.OrphansData{
			Orphans: orphans,
			Count:   len(orphans),
		})

	if len(orphans) == 0 {
		resp.WithInfo("No orphaned files found")
	} else {
		resp.WithWarning("%d orphaned files found", len(orphans))
	}

	writer.Write(resp.Build())
	return nil
}
