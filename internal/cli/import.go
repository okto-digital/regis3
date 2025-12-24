package cli

import (
	"fmt"

	"github.com/okto-digital/regis3/internal/importer"
	"github.com/okto-digital/regis3/internal/output"
	"github.com/spf13/cobra"
)

var importList bool

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Process files in import staging directory",
	Long: `Processes files in the import/ staging directory.

Files that now have valid regis3 frontmatter are moved to their proper
location in the registry. Files still without frontmatter remain in staging.

Use --list to see files pending in the staging directory.

Examples:
  regis3 import          # Process staging directory
  regis3 import --list   # List pending files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if importList {
			return runImportList()
		}
		return runImport()
	},
}

func init() {
	importCmd.Flags().BoolVar(&importList, "list", false, "List pending files")
	rootCmd.AddCommand(importCmd)
}

func runImportList() error {
	debugf("Listing pending imports from: %s", getRegistryPath())

	imp := importer.NewImporter(getRegistryPath())

	if !imp.StagingExists() {
		resp := output.NewResponseBuilder("import").
			WithSuccess(true).
			WithData(output.ImportData{}).
			WithInfo("No files pending in import staging")
		writer.Write(resp.Build())
		return nil
	}

	pending, err := imp.ListPending()
	if err != nil {
		writer.Error(fmt.Sprintf("Failed to list pending: %s", err.Error()))
		return err
	}

	pendingItems := make([]output.PendingItem, len(pending))
	for i, p := range pending {
		pendingItems[i] = output.PendingItem{
			Path:          p.Path,
			SuggestedType: p.SuggestedType,
			SuggestedName: p.SuggestedName,
			Confidence:    p.Confidence,
		}
	}

	resp := output.NewResponseBuilder("import").
		WithSuccess(true).
		WithData(output.ImportData{
			Pending: pendingItems,
		})

	if len(pending) == 0 {
		resp.WithInfo("No files pending")
	} else {
		resp.WithInfo("%d files pending (need regis3 frontmatter)", len(pending))
	}

	writer.Write(resp.Build())
	return nil
}

func runImport() error {
	debugf("Processing import staging from: %s", getRegistryPath())

	imp := importer.NewImporter(getRegistryPath())

	if !imp.StagingExists() {
		resp := output.NewResponseBuilder("import").
			WithSuccess(true).
			WithData(output.ImportData{}).
			WithInfo("No files pending in import staging")
		writer.Write(resp.Build())
		return nil
	}

	result, err := imp.ProcessStaging()
	if err != nil {
		writer.Error(fmt.Sprintf("Import failed: %s", err.Error()))
		return err
	}

	// Build response data
	processed := make([]output.ImportedItem, len(result.Processed))
	for i, p := range result.Processed {
		processed[i] = output.ImportedItem{
			SourcePath: p.SourcePath,
			DestPath:   p.DestPath,
			Type:       p.Type,
			Name:       p.Name,
		}
	}

	pending := make([]output.PendingItem, len(result.Pending))
	for i, p := range result.Pending {
		pending[i] = output.PendingItem{
			Path:          p.Path,
			SuggestedType: p.SuggestedType,
			SuggestedName: p.SuggestedName,
			Confidence:    p.Confidence,
		}
	}

	var errors []string
	for _, e := range result.Errors {
		errors = append(errors, e.Error())
	}

	resp := output.NewResponseBuilder("import").
		WithSuccess(len(result.Errors) == 0).
		WithData(output.ImportData{
			Processed: processed,
			Pending:   pending,
			Errors:    errors,
		})

	if len(processed) > 0 {
		resp.WithInfo("Moved %d files to registry", len(processed))
	}
	if len(pending) > 0 {
		resp.WithInfo("%d files still pending (need regis3 frontmatter)", len(pending))
	}

	for _, e := range errors {
		resp.WithError("import", e)
	}

	writer.Write(resp.Build())

	if len(result.Errors) > 0 {
		return errImportFailed
	}
	return nil
}

var errImportFailed = &exitError{code: 1, message: "import had errors"}
