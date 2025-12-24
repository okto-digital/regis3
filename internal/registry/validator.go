package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Severity indicates the severity of a validation issue.
type Severity int

const (
	SeverityError   Severity = iota // Build fails
	SeverityWarning                 // Build succeeds, issue reported
	SeverityInfo                    // Informational only
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}

// ValidationIssue represents a single validation problem.
type ValidationIssue struct {
	Severity Severity
	Path     string
	Field    string
	Message  string
}

func (v ValidationIssue) String() string {
	if v.Field != "" {
		return fmt.Sprintf("[%s] %s: %s - %s", v.Severity, v.Path, v.Field, v.Message)
	}
	return fmt.Sprintf("[%s] %s: %s", v.Severity, v.Path, v.Message)
}

// ValidationResult contains all validation issues.
type ValidationResult struct {
	Issues []ValidationIssue
}

// HasErrors returns true if there are any error-level issues.
func (r *ValidationResult) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Errors returns only error-level issues.
func (r *ValidationResult) Errors() []ValidationIssue {
	var errors []ValidationIssue
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			errors = append(errors, issue)
		}
	}
	return errors
}

// Warnings returns only warning-level issues.
func (r *ValidationResult) Warnings() []ValidationIssue {
	var warnings []ValidationIssue
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			warnings = append(warnings, issue)
		}
	}
	return warnings
}

// AddError adds an error-level issue.
func (r *ValidationResult) AddError(path, field, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Severity: SeverityError,
		Path:     path,
		Field:    field,
		Message:  message,
	})
}

// AddWarning adds a warning-level issue.
func (r *ValidationResult) AddWarning(path, field, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Severity: SeverityWarning,
		Path:     path,
		Field:    field,
		Message:  message,
	})
}

// AddInfo adds an info-level issue.
func (r *ValidationResult) AddInfo(path, field, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Severity: SeverityInfo,
		Path:     path,
		Field:    field,
		Message:  message,
	})
}

// Validator validates registry items.
type Validator struct {
	// RegistryRoot is the path to the registry root directory.
	RegistryRoot string
}

// NewValidator creates a new validator.
func NewValidator(registryRoot string) *Validator {
	return &Validator{RegistryRoot: registryRoot}
}

// ValidateItems validates a list of items and checks for cross-item issues.
func (v *Validator) ValidateItems(items []*Item) *ValidationResult {
	result := &ValidationResult{}

	// Track seen names for uniqueness check
	seen := make(map[string]string) // fullName -> source path

	for _, item := range items {
		// Validate individual item
		v.validateItem(item, result)

		// Check for duplicate names
		fullName := item.FullName()
		if existingPath, exists := seen[fullName]; exists {
			result.AddError(item.Source, "", fmt.Sprintf("duplicate item '%s' (also defined in %s)", fullName, existingPath))
		} else {
			seen[fullName] = item.Source
		}
	}

	// Validate dependencies exist
	v.validateDependencies(items, seen, result)

	return result
}

// validateItem validates a single item.
func (v *Validator) validateItem(item *Item, result *ValidationResult) {
	// Required: type
	if item.Type == "" {
		result.AddError(item.Source, "type", "required field is missing")
	} else if !IsValidType(item.Type) {
		result.AddError(item.Source, "type", fmt.Sprintf("invalid type '%s' (must be one of: %s)", item.Type, strings.Join(validTypeStrings(), ", ")))
	}

	// Required: name
	if item.Name == "" {
		result.AddError(item.Source, "name", "required field is missing")
	} else {
		// Validate name format (kebab-case)
		if !isKebabCase(item.Name) {
			result.AddWarning(item.Source, "name", "should be kebab-case (lowercase with hyphens)")
		}
	}

	// Required: desc
	if item.Desc == "" {
		result.AddError(item.Source, "desc", "required field is missing")
	} else {
		// Warn if description is too short
		wordCount := len(strings.Fields(item.Desc))
		if wordCount < 3 {
			result.AddWarning(item.Source, "desc", fmt.Sprintf("description is very short (%d words, recommend 10-20)", wordCount))
		}
	}

	// Validate files exist (if specified)
	for _, file := range item.Files {
		filePath := filepath.Join(v.RegistryRoot, item.SourceDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			result.AddError(item.Source, "files", fmt.Sprintf("referenced file does not exist: %s", file))
		}
	}

	// Warn if no tags
	if len(item.Tags) == 0 {
		result.AddWarning(item.Source, "tags", "no tags specified (recommended for searchability)")
	}

	// Validate status if specified
	if item.Status != "" {
		validStatuses := []string{"draft", "stable", "deprecated"}
		isValid := false
		for _, s := range validStatuses {
			if item.Status == s {
				isValid = true
				break
			}
		}
		if !isValid {
			result.AddWarning(item.Source, "status", fmt.Sprintf("unknown status '%s' (expected: draft, stable, deprecated)", item.Status))
		}
	}

	// Validate order for merge types
	if ItemType(item.Type).IsMergeType() && item.Order == 0 {
		result.AddWarning(item.Source, "order", "merge type without order specified (will use default ordering)")
	}

	// Validate trigger for hook type
	if item.Type == string(TypeHook) {
		if item.Trigger == "" {
			result.AddError(item.Source, "trigger", "hook type requires trigger field")
		}
		if item.Run == "" {
			result.AddError(item.Source, "run", "hook type requires run field")
		}
	}

	// Stack type should have dependencies
	if item.Type == string(TypeStack) && len(item.Deps) == 0 {
		result.AddWarning(item.Source, "deps", "stack type should have dependencies")
	}
}

// validateDependencies checks that all referenced dependencies exist.
func (v *Validator) validateDependencies(items []*Item, seen map[string]string, result *ValidationResult) {
	for _, item := range items {
		for _, dep := range item.Deps {
			if _, exists := seen[dep]; !exists {
				result.AddError(item.Source, "deps", fmt.Sprintf("dependency not found: %s", dep))
			}
		}
	}
}

// ValidateItem validates a single item (for use during scanning).
func (v *Validator) ValidateItem(item *Item) *ValidationResult {
	result := &ValidationResult{}
	v.validateItem(item, result)
	return result
}

// validTypeStrings returns valid types as strings for error messages.
func validTypeStrings() []string {
	result := make([]string, len(ValidTypes))
	for i, t := range ValidTypes {
		result[i] = string(t)
	}
	return result
}

// isKebabCase checks if a string is kebab-case.
func isKebabCase(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	// Should not start or end with hyphen
	if s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	// Should not have consecutive hyphens
	if strings.Contains(s, "--") {
		return false
	}
	return true
}
