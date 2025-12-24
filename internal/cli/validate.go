package cli

import (
	"fmt"

	"github.com/okto-digital/regis3/internal/output"
	"github.com/okto-digital/regis3/internal/registry"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the registry",
	Long: `Validates all items in the registry for:
- Required fields (type, name, desc)
- Valid type values
- Unique type:name combinations
- Existing dependencies
- File references`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runValidate()
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate() error {
	debugf("Validating registry: %s", getRegistryPath())

	// Build and validate
	result, err := registry.BuildRegistry(getRegistryPath())
	if err != nil {
		writer.Error(fmt.Sprintf("Failed to scan registry: %s", err.Error()))
		return err
	}

	itemCount := len(result.Manifest.Items)
	issues := result.Validation.Issues

	// Build response
	resp := output.NewResponseBuilder("validate").
		WithData(output.ValidateData{
			ItemCount:  itemCount,
			ErrorCount: countSeverity(issues, registry.SeverityError),
			WarnCount:  countSeverity(issues, registry.SeverityWarning),
			InfoCount:  countSeverity(issues, registry.SeverityInfo),
		})

	// Add issues as messages
	hasErrors := false
	for _, issue := range issues {
		switch issue.Severity {
		case registry.SeverityError:
			resp.WithError(issue.Path, issue.Message)
			hasErrors = true
		case registry.SeverityWarning:
			resp.WithWarning("%s: %s", issue.Path, issue.Message)
		case registry.SeverityInfo:
			resp.WithInfo("%s: %s", issue.Path, issue.Message)
		}
	}

	if hasErrors {
		resp.WithSuccess(false)
	} else {
		resp.WithSuccess(true)
		if len(issues) == 0 {
			resp.WithInfo("All %d items are valid", itemCount)
		}
	}

	writer.Write(resp.Build())

	if hasErrors {
		return errValidationFailed
	}
	return nil
}

func countSeverity(issues []registry.ValidationIssue, severity registry.Severity) int {
	count := 0
	for _, issue := range issues {
		if issue.Severity == severity {
			count++
		}
	}
	return count
}

var errValidationFailed = &exitError{code: 1, message: "validation failed"}

type exitError struct {
	code    int
	message string
}

func (e *exitError) Error() string {
	return e.message
}
