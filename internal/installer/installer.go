package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/okto-digital/regis3/internal/registry"
	"github.com/okto-digital/regis3/internal/resolver"
)

// Installer handles installing registry items to a project.
type Installer struct {
	// Target is the installation target configuration.
	Target *Target

	// ProjectDir is the project directory to install to.
	ProjectDir string

	// RegistryPath is the path to the registry.
	RegistryPath string

	// Tracker tracks installed items.
	Tracker *Tracker

	// Transformer handles content transformations.
	Transformer *Transformer

	// DryRun if true, only simulates installation.
	DryRun bool

	// Force if true, reinstalls even if up to date.
	Force bool
}

// NewInstaller creates a new installer.
func NewInstaller(projectDir, registryPath string, target *Target) (*Installer, error) {
	tracker, err := LoadTracker(projectDir, target.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to load tracker: %w", err)
	}

	tracker.SetRegistryPath(registryPath)

	return &Installer{
		Target:       target,
		ProjectDir:   projectDir,
		RegistryPath: registryPath,
		Tracker:      tracker,
		Transformer:  NewTransformer(target),
		DryRun:       false,
		Force:        false,
	}, nil
}

// InstallResult contains the result of an installation operation.
type InstallResult struct {
	// Installed are items that were installed.
	Installed []string

	// Updated are items that were updated.
	Updated []string

	// Skipped are items that were already up to date.
	Skipped []string

	// Errors are installation errors.
	Errors []InstallError

	// MergedItems are items merged into CLAUDE.md.
	MergedItems []string
}

// InstallError represents an installation error.
type InstallError struct {
	ItemID  string
	Message string
	Err     error
}

func (e InstallError) Error() string {
	return fmt.Sprintf("%s: %s", e.ItemID, e.Message)
}

// Install installs the specified items and their dependencies.
func (i *Installer) Install(manifest *registry.Manifest, itemIDs []string) (*InstallResult, error) {
	result := &InstallResult{}

	// Resolve dependencies
	r := resolver.NewResolver(manifest)
	resolved, err := r.Resolve(itemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Check for missing dependencies
	if len(resolved.Missing) > 0 {
		return nil, fmt.Errorf("missing dependencies: %v", resolved.Missing)
	}

	// Prepare merge content
	mergeContent := NewMergeContent()

	// Install each item in order
	for _, item := range resolved.Items {
		itemResult, err := i.installItem(item, mergeContent)
		if err != nil {
			result.Errors = append(result.Errors, InstallError{
				ItemID:  item.FullName(),
				Message: err.Error(),
				Err:     err,
			})
			continue
		}

		switch itemResult {
		case installResultInstalled:
			result.Installed = append(result.Installed, item.FullName())
		case installResultUpdated:
			result.Updated = append(result.Updated, item.FullName())
		case installResultSkipped:
			result.Skipped = append(result.Skipped, item.FullName())
		case installResultMerged:
			result.MergedItems = append(result.MergedItems, item.FullName())
		}
	}

	// Write merged content to CLAUDE.md
	if mergeContent.HasContent() {
		if err := i.writeMergeFile(mergeContent); err != nil {
			result.Errors = append(result.Errors, InstallError{
				ItemID:  "CLAUDE.md",
				Message: err.Error(),
				Err:     err,
			})
		}
	}

	// Save tracker
	if !i.DryRun {
		if err := i.Tracker.Save(); err != nil {
			return result, fmt.Errorf("failed to save tracker: %w", err)
		}
	}

	return result, nil
}

type installResultType int

const (
	installResultInstalled installResultType = iota
	installResultUpdated
	installResultSkipped
	installResultMerged
)

// installItem installs a single item.
func (i *Installer) installItem(item *registry.Item, mergeContent *MergeContent) (installResultType, error) {
	// Transform content
	content, err := i.Transformer.Transform(item)
	if err != nil {
		return 0, fmt.Errorf("failed to transform content: %w", err)
	}

	// Calculate content hash
	hash := hashContent(content)

	// Check if needs update
	if !i.Force && !i.Tracker.NeedsUpdate(item.FullName(), hash) {
		return installResultSkipped, nil
	}

	// Handle merge types
	if i.Target.IsMergeType(item.Type) {
		mergeContent.Add(item, content)
		if !i.DryRun {
			i.Tracker.MarkInstalled(item.FullName(), item.Type, item.Name, i.Target.MergeFile, true)
			i.Tracker.SetSourceHash(item.FullName(), hash)
		}
		return installResultMerged, nil
	}

	// Handle stack type (meta-type, no direct installation)
	if item.Type == "stack" {
		// Stack is just a dependency grouping, nothing to install
		if !i.DryRun {
			i.Tracker.MarkInstalled(item.FullName(), item.Type, item.Name, "", false)
			i.Tracker.SetSourceHash(item.FullName(), hash)
		}
		return installResultSkipped, nil
	}

	// Get installation path
	destPath, err := i.Target.GetPath(item.Type, item.Name)
	if err != nil {
		return 0, fmt.Errorf("failed to get installation path: %w", err)
	}

	fullPath := filepath.Join(i.ProjectDir, destPath)

	// Check if already installed
	isUpdate := i.Tracker.IsInstalled(item.FullName())

	// Write file
	if !i.DryRun {
		if err := i.writeFile(fullPath, content); err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}

		// Copy additional files if specified
		if len(item.Files) > 0 {
			if err := i.copyAdditionalFiles(item, filepath.Dir(fullPath)); err != nil {
				return 0, fmt.Errorf("failed to copy additional files: %w", err)
			}
		}

		// Update tracker
		i.Tracker.MarkInstalled(item.FullName(), item.Type, item.Name, destPath, false)
		i.Tracker.SetSourceHash(item.FullName(), hash)
	}

	if isUpdate {
		return installResultUpdated, nil
	}
	return installResultInstalled, nil
}

// writeFile writes content to a file, creating directories as needed.
func (i *Installer) writeFile(path, content string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// copyAdditionalFiles copies additional files specified in the item.
func (i *Installer) copyAdditionalFiles(item *registry.Item, destDir string) error {
	for _, file := range item.Files {
		srcPath := filepath.Join(i.RegistryPath, item.SourceDir, file)
		destPath := filepath.Join(destDir, file)

		// Read source file
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Create destination directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file, err)
		}

		// Write destination file
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", file, err)
		}
	}

	return nil
}

// writeMergeFile writes merged content to CLAUDE.md.
func (i *Installer) writeMergeFile(mergeContent *MergeContent) error {
	mergeFilePath := filepath.Join(i.ProjectDir, i.Target.MergeFile)

	// Read existing file if it exists
	existing := ""
	if data, err := os.ReadFile(mergeFilePath); err == nil {
		existing = string(data)
	}

	// Generate new merged content
	newContent := mergeContent.Generate()

	// Update or create file
	finalContent := UpdateExistingFile(existing, newContent)

	if i.DryRun {
		return nil
	}

	return os.WriteFile(mergeFilePath, []byte(finalContent), 0644)
}

// Uninstall removes installed items.
func (i *Installer) Uninstall(itemIDs []string) (*UninstallResult, error) {
	result := &UninstallResult{}

	for _, id := range itemIDs {
		installed := i.Tracker.GetInstalled(id)
		if installed == nil {
			result.NotFound = append(result.NotFound, id)
			continue
		}

		// Skip merge types for now (would need to regenerate CLAUDE.md)
		if installed.Merged {
			result.Skipped = append(result.Skipped, id)
			continue
		}

		// Delete the file if it exists and has a path
		if installed.InstalledPath != "" {
			fullPath := filepath.Join(i.ProjectDir, installed.InstalledPath)
			if !i.DryRun {
				// Delete file
				if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
					result.Errors = append(result.Errors, InstallError{
						ItemID:  id,
						Message: err.Error(),
						Err:     err,
					})
					continue
				}

				// Try to remove empty parent directory
				dir := filepath.Dir(fullPath)
				os.Remove(dir) // Ignore error if not empty
			}
		}

		if !i.DryRun {
			i.Tracker.MarkUninstalled(id)
		}
		result.Uninstalled = append(result.Uninstalled, id)
	}

	// Save tracker
	if !i.DryRun {
		if err := i.Tracker.Save(); err != nil {
			return result, fmt.Errorf("failed to save tracker: %w", err)
		}
	}

	return result, nil
}

// UninstallResult contains the result of an uninstall operation.
type UninstallResult struct {
	Uninstalled []string
	Skipped     []string
	NotFound    []string
	Errors      []InstallError
}

// Status returns the installation status for items.
func (i *Installer) Status(manifest *registry.Manifest) *StatusResult {
	result := &StatusResult{
		Items: make(map[string]*ItemStatus),
	}

	for id, item := range manifest.Items {
		status := &ItemStatus{
			ID:   id,
			Type: item.Type,
			Name: item.Name,
		}

		installed := i.Tracker.GetInstalled(id)
		if installed != nil {
			status.Installed = true
			status.InstalledAt = installed.InstalledAt
			status.UpdatedAt = installed.UpdatedAt
			status.Path = installed.InstalledPath
			status.Merged = installed.Merged

			// Check if needs update
			content, _ := i.Transformer.Transform(item)
			hash := hashContent(content)
			status.NeedsUpdate = installed.SourceHash != hash
		}

		result.Items[id] = status
	}

	return result
}

// StatusResult contains installation status for items.
type StatusResult struct {
	Items map[string]*ItemStatus
}

// ItemStatus contains status for a single item.
type ItemStatus struct {
	ID          string
	Type        string
	Name        string
	Installed   bool
	InstalledAt interface{}
	UpdatedAt   interface{}
	Path        string
	Merged      bool
	NeedsUpdate bool
}

// hashContent returns a SHA256 hash of content.
func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}
