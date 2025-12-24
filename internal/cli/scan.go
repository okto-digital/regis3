package cli

import (
	"fmt"

	"github.com/okto-digital/regis3/internal/importer"
	"github.com/okto-digital/regis3/internal/output"
	"github.com/spf13/cobra"
)

var scanDryRun bool

var scanCmd = &cobra.Command{
	Use:   "scan <path>",
	Short: "Scan external path for markdown files",
	Long: `Scans an external directory for markdown files and imports them to the registry.

Files with valid regis3 frontmatter are imported directly to the appropriate
directory. Files without regis3 frontmatter are placed in the import/
staging directory for manual review.

Examples:
  regis3 scan ~/Documents/prompts
  regis3 scan ./my-skills --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScan(args[0])
	},
}

func init() {
	scanCmd.Flags().BoolVar(&scanDryRun, "dry-run", false, "Preview what would be imported")
	rootCmd.AddCommand(scanCmd)
}

func runScan(path string) error {
	debugf("Scanning: %s", path)

	imp := importer.NewImporter(getRegistryPath())
	imp.DryRun = scanDryRun

	result, err := imp.ScanAndImport(path)
	if err != nil {
		writer.Error(fmt.Sprintf("Scan failed: %s", err.Error()))
		return err
	}

	// Build response data
	imported := make([]output.ImportedItem, len(result.Imported))
	for i, item := range result.Imported {
		imported[i] = output.ImportedItem{
			SourcePath: item.SourcePath,
			DestPath:   item.DestPath,
			Type:       item.Type,
			Name:       item.Name,
		}
	}

	staged := make([]output.ImportedItem, len(result.Staged))
	for i, item := range result.Staged {
		staged[i] = output.ImportedItem{
			SourcePath: item.SourcePath,
			DestPath:   item.DestPath,
			Type:       item.Type,
			Name:       item.Name,
		}
	}

	var errors []string
	for _, e := range result.Errors {
		errors = append(errors, e.Error())
	}

	resp := output.NewResponseBuilder("scan").
		WithSuccess(len(result.Errors) == 0).
		WithData(output.ScanData{
			Imported: imported,
			Staged:   staged,
			Errors:   errors,
			DryRun:   scanDryRun,
		})

	if scanDryRun {
		resp.WithInfo("Would import %d files, stage %d files (dry run)", len(imported), len(staged))
	} else {
		if len(imported) > 0 {
			resp.WithInfo("Imported %d files to registry", len(imported))
		}
		if len(staged) > 0 {
			resp.WithInfo("Staged %d files in import/ (need regis3 headers)", len(staged))
		}
	}

	for _, e := range errors {
		resp.WithError("scan", e)
	}

	writer.Write(resp.Build())

	if len(result.Errors) > 0 {
		return errScanFailed
	}
	return nil
}

var errScanFailed = &exitError{code: 1, message: "scan had errors"}
