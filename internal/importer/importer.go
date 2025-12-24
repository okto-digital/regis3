package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/okto-digital/regis3/internal/registry"
)

const (
	// ImportDir is the staging directory for files pending YAML headers.
	ImportDir = "import"
)

// Importer handles importing external files into the registry.
type Importer struct {
	// RegistryPath is the path to the registry.
	RegistryPath string

	// Scanner scans external paths.
	Scanner *ExternalScanner

	// Classifier classifies files.
	Classifier *Classifier

	// DryRun if true, only simulates import.
	DryRun bool
}

// NewImporter creates a new importer.
func NewImporter(registryPath string) *Importer {
	return &Importer{
		RegistryPath: registryPath,
		Scanner:      NewExternalScanner(),
		Classifier:   NewClassifier(),
		DryRun:       false,
	}
}

// ImportResult contains the result of an import operation.
type ImportResult struct {
	// Imported are files copied directly to the registry (had valid regis3).
	Imported []ImportedFile

	// Staged are files copied to the import/ staging directory.
	Staged []ImportedFile

	// Skipped are files that were skipped (already exist, etc.).
	Skipped []SkippedFile

	// Errors are import errors.
	Errors []ImportError
}

// ImportedFile represents an imported file.
type ImportedFile struct {
	// SourcePath is the original file path.
	SourcePath string

	// DestPath is where the file was copied to.
	DestPath string

	// Type is the regis3 type.
	Type string

	// Name is the regis3 name.
	Name string

	// WasStaged indicates if the file was staged (no regis3 block).
	WasStaged bool
}

// SkippedFile represents a skipped file.
type SkippedFile struct {
	Path   string
	Reason string
}

// ImportError represents an import error.
type ImportError struct {
	Path    string
	Message string
	Err     error
}

func (e ImportError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// ScanAndImport scans a path and imports found files.
func (i *Importer) ScanAndImport(externalPath string) (*ImportResult, error) {
	// Scan the external path
	scanResult, err := i.Scanner.Scan(externalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan: %w", err)
	}

	result := &ImportResult{}

	// Copy scan errors to result
	for _, scanErr := range scanResult.Errors {
		result.Errors = append(result.Errors, ImportError{
			Path:    scanErr.Path,
			Message: scanErr.Message,
			Err:     scanErr.Err,
		})
	}

	// Process each file
	for _, file := range scanResult.Files {
		imported, err := i.importFile(file)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Path:    file.Path,
				Message: err.Error(),
				Err:     err,
			})
			continue
		}

		if imported == nil {
			// Skipped
			continue
		}

		if imported.WasStaged {
			result.Staged = append(result.Staged, *imported)
		} else {
			result.Imported = append(result.Imported, *imported)
		}
	}

	return result, nil
}

// importFile imports a single file.
func (i *Importer) importFile(file ScannedFile) (*ImportedFile, error) {
	// Classify the file
	class, err := i.Classifier.Classify(file.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to classify: %w", err)
	}

	// Determine destination
	var destPath string
	var wasStaged bool

	if class.HasValidRegis3 {
		// Has valid regis3 - import directly to registry
		destPath = i.getRegistryPath(class.ExistingMeta.Type, class.ExistingMeta.Name)
	} else {
		// No regis3 - stage in import/ directory
		destPath = filepath.Join(i.RegistryPath, ImportDir, file.RelPath)
		wasStaged = true
	}

	// Check if destination already exists
	if !i.DryRun {
		if _, err := os.Stat(destPath); err == nil {
			// File exists - skip
			return nil, nil
		}
	}

	// Copy file
	if !i.DryRun {
		if err := i.copyFile(file.Path, destPath); err != nil {
			return nil, fmt.Errorf("failed to copy: %w", err)
		}
	}

	typeName := class.SuggestedType
	name := class.SuggestedName
	if class.HasValidRegis3 {
		typeName = class.ExistingMeta.Type
		name = class.ExistingMeta.Name
	}

	return &ImportedFile{
		SourcePath: file.Path,
		DestPath:   destPath,
		Type:       typeName,
		Name:       name,
		WasStaged:  wasStaged,
	}, nil
}

// getRegistryPath returns the path in the registry for an item type.
func (i *Importer) getRegistryPath(itemType, name string) string {
	// Map types to directories
	dirMap := map[string]string{
		"skill":      "skills",
		"subagent":   "agents",
		"philosophy": "philosophies",
		"command":    "commands",
		"mcp":        "mcp",
		"script":     "scripts",
		"doc":        "docs",
		"project":    "projects",
		"ruleset":    "rulesets",
		"stack":      "stacks",
		"hook":       "hooks",
		"prompt":     "prompts",
	}

	dir := dirMap[itemType]
	if dir == "" {
		dir = itemType + "s"
	}

	return filepath.Join(i.RegistryPath, dir, name+".md")
}

// copyFile copies a file from src to dest.
func (i *Importer) copyFile(src, dest string) error {
	// Read source
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Create destination directory
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write destination
	return os.WriteFile(dest, content, 0644)
}

// ProcessStaging processes files in the import/ staging directory.
func (i *Importer) ProcessStaging() (*ProcessResult, error) {
	stagingDir := filepath.Join(i.RegistryPath, ImportDir)

	// Check if staging directory exists
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		return &ProcessResult{}, nil
	}

	result := &ProcessResult{}

	// Walk the staging directory
	err := filepath.Walk(stagingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" {
			return nil
		}

		// Classify the file
		class, err := i.Classifier.Classify(path)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{
				Path:    path,
				Message: err.Error(),
				Err:     err,
			})
			return nil
		}

		if !class.HasValidRegis3 {
			// Still no regis3 block - add to pending
			result.Pending = append(result.Pending, PendingFile{
				Path:          path,
				SuggestedType: class.SuggestedType,
				SuggestedName: class.SuggestedName,
				Confidence:    class.Confidence,
				Reason:        class.Reason,
			})
			return nil
		}

		// Has regis3 - move to proper location
		destPath := i.getRegistryPath(class.ExistingMeta.Type, class.ExistingMeta.Name)

		if !i.DryRun {
			// Copy to new location
			if err := i.copyFile(path, destPath); err != nil {
				result.Errors = append(result.Errors, ImportError{
					Path:    path,
					Message: "failed to copy: " + err.Error(),
					Err:     err,
				})
				return nil
			}

			// Remove from staging
			if err := os.Remove(path); err != nil {
				result.Errors = append(result.Errors, ImportError{
					Path:    path,
					Message: "failed to remove staged file: " + err.Error(),
					Err:     err,
				})
			}
		}

		result.Processed = append(result.Processed, ProcessedFile{
			SourcePath: path,
			DestPath:   destPath,
			Type:       class.ExistingMeta.Type,
			Name:       class.ExistingMeta.Name,
		})

		return nil
	})

	return result, err
}

// ProcessResult contains the result of processing the staging directory.
type ProcessResult struct {
	// Processed are files that were moved to the registry.
	Processed []ProcessedFile

	// Pending are files still waiting for regis3 headers.
	Pending []PendingFile

	// Errors are processing errors.
	Errors []ImportError
}

// ProcessedFile represents a file that was processed from staging.
type ProcessedFile struct {
	SourcePath string
	DestPath   string
	Type       string
	Name       string
}

// PendingFile represents a file still pending in staging.
type PendingFile struct {
	Path          string
	SuggestedType string
	SuggestedName string
	Confidence    int
	Reason        string
}

// ListPending lists files pending in the staging directory.
func (i *Importer) ListPending() ([]PendingFile, error) {
	stagingDir := filepath.Join(i.RegistryPath, ImportDir)

	// Check if staging directory exists
	if _, err := os.Stat(stagingDir); os.IsNotExist(err) {
		return nil, nil
	}

	var pending []PendingFile

	err := filepath.Walk(stagingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" {
			return nil
		}

		// Classify
		class, err := i.Classifier.Classify(path)
		if err != nil {
			return nil
		}

		if !class.HasValidRegis3 {
			relPath, _ := filepath.Rel(stagingDir, path)
			pending = append(pending, PendingFile{
				Path:          relPath,
				SuggestedType: class.SuggestedType,
				SuggestedName: class.SuggestedName,
				Confidence:    class.Confidence,
				Reason:        class.Reason,
			})
		}

		return nil
	})

	return pending, err
}

// Reindex rebuilds the registry manifest.
func (i *Importer) Reindex() (*registry.BuildResult, error) {
	return registry.BuildRegistry(i.RegistryPath)
}

// StagingExists checks if the staging directory has files.
func (i *Importer) StagingExists() bool {
	stagingDir := filepath.Join(i.RegistryPath, ImportDir)
	info, err := os.Stat(stagingDir)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}

	// Check if directory has files
	entries, err := os.ReadDir(stagingDir)
	if err != nil {
		return false
	}

	return len(entries) > 0
}
